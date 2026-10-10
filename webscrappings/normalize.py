"""source_records -> items / item_translations / external_scores.

Every record of a QID is re-read whenever any of them changes, and the item is rebuilt from scratch by
`build_item` (pure), so the result never depends on the order sources arrived in.
"""

import logging
import re
import unicodedata
import uuid
from dataclasses import dataclass, field
from datetime import date
from typing import Any

import psycopg
from psycopg.types.json import Jsonb
from pydantic import BaseModel, ValidationError

from common.cli import setup_logging
from common.db import busy, connect
from connectors.wikidata import Claim, WikidataPayload
from revalidate import revalidate

log = logging.getLogger("normalize")

SOURCES = [
    "wikidata",
    "wikipedia:en",
    "wikipedia:pt",
    "wikipedia:es",
    "openlibrary",
    "steam",
    "gog",
    "tmdb",
    "igdb",
    "synopsis",
]
BATCH = 200
SYNOPSIS_LICENSE = "CC BY-SA 4.0"

EXTERNAL_ID_PROPS = {
    "imdb": "P345",
    "igdb": "P5794",
    "steam": "P1733",
    "isbn13": "P212",
    "openlibrary": "P648",
    "hltb": "P2816",
    "goodreads": "P8383",
    "metacritic": "P1712",
    "rottentomatoes": "P1258",
    "epic": "P6278",
    "gog": "P2725",
}
TMDB_PROP = {"movie": "P4947", "series": "P4983"}
CREATOR_PROP = {"movie": "P57", "series": "P170", "game": "P178", "book": "P50"}
# locale -> (Wikidata label languages in order of preference, Wikipedia language)
LOCALES = {"en": (("en", "mul"), "en"), "pt-BR": (("pt-br", "pt", "mul"), "pt"), "es": (("es", "mul"), "es")}
MINUTES_PER_UNIT = {"Q7727": 1.0, "Q25235": 60.0, "Q11574": 1 / 60, "1": 1.0}

METACRITIC, ROTTEN_TOMATOES = "Q150248", "Q105584"
METASCORE, TOMATOMETER = "Q106515043", "Q108403393"
SCORE_SOURCES = ["metacritic", "rottentomatoes"]
COVER_KEYS = ("coverUrl", "coverThumbUrl", "coverAttribution")


class WikipediaPayload(BaseModel):
    title: str
    extract: str
    url: str


class OpenLibraryPayload(BaseModel):
    pages: int | None = None


class PlatformsPayload(BaseModel):
    platforms: list[str] = []


class EnrichmentPayload(BaseModel):
    """Shared shape of the TMDB and IGDB payloads. Every field is a fallback: Wikidata wins when it has a value."""

    id: int | None = None
    genres: list[str] = []
    platforms: list[str] = []
    creators: list[str] = []
    countries: list[str] = []
    trailers: list[str] = []
    runtime: int | None = None
    seasons: int | None = None
    episodes: int | None = None
    release_date: date | None = None


def enrichment(records: dict[str, dict]) -> EnrichmentPayload:
    # TMDB only has movie/series records and IGDB only games, so at most one of them exists per item.
    return EnrichmentPayload.model_validate(records.get("tmdb") or records.get("igdb") or {})


@dataclass(frozen=True)
class Translation:
    locale: str
    title: str
    synopsis: str | None = None
    synopsis_license: str | None = None
    synopsis_source_url: str | None = None


@dataclass(frozen=True)
class Score:
    source: str
    value: str
    url: str | None


@dataclass
class NormalizedItem:
    qid: str
    type: str
    release_date: date | None
    external_ids: dict[str, str]
    meta: dict[str, Any]
    popularity: float
    translations: list[Translation] = field(default_factory=list)
    scores: list[Score] = field(default_factory=list)

    @property
    def title_en(self) -> str:
        return next(t.title for t in self.translations if t.locale == "en")


def slugify(text: str) -> str:
    ascii_text = unicodedata.normalize("NFKD", text).encode("ascii", "ignore").decode()
    return re.sub(r"[^a-z0-9]+", "-", ascii_text.lower()).strip("-")[:80].rstrip("-")


def slug_candidates(title: str, year: int | None, qid: str) -> list[str]:
    base = slugify(title)
    if not base:
        return [qid.lower()]
    candidates = [base]
    if year:
        candidates.append(f"{base}-{year}")
    candidates.append(f"{base}-{qid.lower()}")
    return candidates


def parse_time(value: Any) -> date | None:
    """Wikidata time -> date, padding month/day for month or year precision; None below year precision."""
    if not isinstance(value, dict) or not isinstance(value.get("time"), str):
        return None
    match = re.match(r"^\+(\d{1,4})-(\d\d)-(\d\d)T", value["time"])
    precision = value.get("precision")
    if not match or not isinstance(precision, int) or precision < 9:
        return None
    year, month, day = (int(g) for g in match.groups())
    month = month if precision >= 10 and month else 1
    day = day if precision >= 11 and day else 1
    try:
        return date(year, month, day)
    except ValueError:
        return None


def _ranked(claims: list[Claim]) -> list[Claim]:
    return sorted(claims, key=lambda c: c.rank != "preferred")


def first_value(wd: WikidataPayload, prop: str) -> Any:
    claims = _ranked(wd.claims.get(prop, []))
    return claims[0].value if claims else None


def quantity(wd: WikidataPayload, prop: str, units: dict[str, float] | None = None) -> int | None:
    value = first_value(wd, prop)
    if not isinstance(value, dict):
        return None
    try:
        amount = float(str(value.get("amount")))
    except ValueError:
        return None
    factor = (units or {"1": 1.0}).get(str(value.get("unit")))
    if factor is None or amount <= 0:
        return None
    return round(amount * factor) or None


def ref_labels(wd: WikidataPayload, prop: str) -> list[str]:
    labels = [wd.refs.get(str(c.value)) for c in _ranked(wd.claims.get(prop, []))]
    return list(dict.fromkeys(label for label in labels if label))


def release_date(wd: WikidataPayload) -> date | None:
    props = ("P580", "P577", "P571") if wd.type == "series" else ("P577", "P580", "P571")
    for prop in props:
        dates = [d for c in wd.claims.get(prop, []) if (d := parse_time(c.value))]
        if dates:
            return min(dates)
    return None


def external_ids(wd: WikidataPayload) -> dict[str, str]:
    props = dict(EXTERNAL_ID_PROPS)
    if wd.type in TMDB_PROP:
        props["tmdb"] = TMDB_PROP[wd.type]
    ids = {key: str(v) for key, prop in props.items() if isinstance(v := first_value(wd, prop), str) and v}
    if "isbn13" in ids:
        ids["isbn13"] = re.sub(r"[^0-9X]", "", ids["isbn13"].upper())
    return ids


def _score_time(claim: Claim) -> str:
    times = [t.get("time", "") for t in claim.qualifiers.get("P585", []) if isinstance(t, dict)]
    return max((str(t) for t in times), default="")


def scores(wd: WikidataPayload, ids: dict[str, str]) -> list[Score]:
    """Metascore and Tomatometer from P444 (review score) qualified by P447 (review score by)."""
    best: dict[str, Claim] = {}
    for claim in wd.claims.get("P444", []):
        by = claim.qualifiers.get("P447", [])
        method = claim.qualifiers.get("P459", [])
        value = str(claim.value)
        if METACRITIC in by and (not method or METASCORE in method) and re.fullmatch(r"\d{1,3}/100", value):
            source = "metacritic"
        elif ROTTEN_TOMATOES in by and (TOMATOMETER in method or (not method and re.fullmatch(r"\d{1,3}%", value))):
            source = "rottentomatoes"
        else:
            continue
        current = best.get(source)
        key = (claim.rank == "preferred", _score_time(claim))
        if current is None or key > (current.rank == "preferred", _score_time(current)):
            best[source] = claim
    urls = {
        "metacritic": f"https://www.metacritic.com/{ids['metacritic']}" if "metacritic" in ids else None,
        "rottentomatoes": f"https://www.rottentomatoes.com/{ids['rottentomatoes']}"
        if "rottentomatoes" in ids
        else None,
    }
    return [Score(source, str(claim.value), urls[source]) for source, claim in sorted(best.items())]


def build_meta(wd: WikidataPayload, records: dict[str, dict]) -> dict[str, Any]:
    extra = enrichment(records)
    # IGDB lists every platform; Steam and GOG only know the PC operating systems.
    platforms = (ref_labels(wd, "P400") or extra.platforms) if wd.type == "game" else []
    for source in ("steam", "gog"):
        if not platforms and source in records:
            platforms = PlatformsPayload.model_validate(records[source]).platforms
    pages = quantity(wd, "P1104") if wd.type == "book" else None
    if wd.type == "book" and not pages and "openlibrary" in records:
        pages = OpenLibraryPayload.model_validate(records["openlibrary"]).pages
    trailers = list(dict.fromkeys(str(c.value) for c in _ranked(wd.claims.get("P1651", []))))[:3]
    screen = wd.type in ("movie", "series")
    meta: dict[str, Any] = {
        "genres": ref_labels(wd, "P136") or extra.genres,
        "platforms": platforms,
        "creators": ref_labels(wd, CREATOR_PROP[wd.type]) or extra.creators,
        "runtimeMinutes": (quantity(wd, "P2047", MINUTES_PER_UNIT) or extra.runtime) if screen else None,
        "pages": pages,
        "seasons": (quantity(wd, "P2437") or extra.seasons) if wd.type == "series" else None,
        "episodes": (quantity(wd, "P1113") or extra.episodes) if wd.type == "series" else None,
        "countries": ref_labels(wd, "P495") or extra.countries,
        "trailers": [{"youtubeId": t} for t in trailers or extra.trailers[:3]],
    }
    return {k: v for k, v in meta.items() if v}


def build_translations(wd: WikidataPayload, records: dict[str, dict]) -> list[Translation]:
    out = []
    for locale, (label_langs, wiki_lang) in LOCALES.items():
        label = next((wd.labels[lang] for lang in label_langs if wd.labels.get(lang)), None)
        raw_summary = records.get(f"wikipedia:{wiki_lang}")
        summary = WikipediaPayload.model_validate(raw_summary) if raw_summary else None
        if locale == "en":
            label = label or wd.sitelinks.get("enwiki")
        title = label or (summary.title if summary else None)
        if not title:
            continue
        if summary:
            out.append(Translation(locale, title, summary.extract, SYNOPSIS_LICENSE, summary.url))
        else:
            # Own text from the synopsis rewriting step (source `synopsis`): no third-party license or source link.
            own = records.get("synopsis", {}).get(locale)
            out.append(Translation(locale, title, own if isinstance(own, str) and own else None))
    return out


def build_item(qid: str, records: dict[str, dict]) -> NormalizedItem | None:
    """Pure mapping from all source payloads of one QID (keyed by source) to canonical rows."""
    if "wikidata" not in records:
        return None
    wd = WikidataPayload.model_validate(records["wikidata"])
    translations = build_translations(wd, records)
    if not any(t.locale == "en" for t in translations):
        return None
    ids = external_ids(wd)
    extra = enrichment(records)
    if "tmdb" in records and extra.id and "tmdb" not in ids:
        # Resolved by the TMDB connector through the IMDb ID.
        ids["tmdb"] = str(extra.id)
    return NormalizedItem(
        qid=qid,
        type=wd.type,
        release_date=release_date(wd) or extra.release_date,
        external_ids=ids,
        meta=build_meta(wd, records),
        popularity=float(wd.sitelinks_count),
        translations=translations,
        scores=scores(wd, ids),
    )


UPSERT_ITEM = """
INSERT INTO items (id, type, slug, wikidata_qid, release_date, external_ids, meta, popularity)
VALUES (%(id)s, %(type)s, %(slug)s, %(qid)s, %(release_date)s, %(external_ids)s, %(meta)s, %(popularity)s)
ON CONFLICT (wikidata_qid) DO UPDATE SET
    release_date = excluded.release_date,
    external_ids = excluded.external_ids,
    -- covers.py owns the cover keys; keep them across re-normalization.
    meta = excluded.meta || jsonb_strip_nulls(jsonb_build_object(
        'coverUrl', items.meta->'coverUrl', 'coverThumbUrl', items.meta->'coverThumbUrl',
        'coverAttribution', items.meta->'coverAttribution')),
    popularity = excluded.popularity,
    updated_at = CASE WHEN items.release_date IS DISTINCT FROM excluded.release_date
        OR items.external_ids <> excluded.external_ids
        OR (items.meta - 'coverUrl' - 'coverThumbUrl' - 'coverAttribution') <> excluded.meta
        THEN now() ELSE items.updated_at END
RETURNING id
"""

UPSERT_TRANSLATION = """
INSERT INTO item_translations (item_id, locale, title, synopsis, synopsis_license, synopsis_source_url)
VALUES (%s, %s, %s, %s, %s, %s)
ON CONFLICT (item_id, locale) DO UPDATE SET title = excluded.title, synopsis = excluded.synopsis,
    synopsis_license = excluded.synopsis_license, synopsis_source_url = excluded.synopsis_source_url
"""

UPSERT_SCORE = """
INSERT INTO external_scores (item_id, source, value, url, fetched_at) VALUES (%s, %s, %s, %s, now())
ON CONFLICT (item_id, source) DO UPDATE SET value = excluded.value, url = excluded.url, fetched_at = now()
"""


def choose_slug(conn: psycopg.Connection, item: NormalizedItem) -> str:
    year = item.release_date.year if item.release_date else None
    for slug in slug_candidates(item.title_en, year, item.qid):
        taken = conn.execute(
            "SELECT 1 FROM items WHERE type = %s AND slug = %s AND wikidata_qid <> %s", (item.type, slug, item.qid)
        ).fetchone()
        if not taken:
            return slug
    raise RuntimeError(f"no free slug for {item.qid}")


def known_items(conn: psycopg.Connection, qids: list[str]) -> dict[str, tuple[uuid.UUID, str]]:
    rows = conn.execute("SELECT wikidata_qid, id, slug FROM items WHERE wikidata_qid = ANY(%s)", (qids,))
    return {r[0]: (r[1], r[2]) for r in rows}


def save_item(
    conn: psycopg.Connection, item: NormalizedItem, known: dict[str, tuple[uuid.UUID, str]] | None = None
) -> uuid.UUID:
    """`known` (from `known_items` for the whole batch) skips the per-item lookup. An existing item keeps its id, so
    nothing here waits for a result and the statements stay queued in the caller's pipeline."""
    existing = known.get(item.qid) if known is not None else known_items(conn, [item.qid]).get(item.qid)
    item_id, slug = existing if existing else (uuid.uuid7(), choose_slug(conn, item))
    conn.execute(
        UPSERT_ITEM,
        {
            "id": item_id,
            "type": item.type,
            "slug": slug,
            "qid": item.qid,
            "release_date": item.release_date,
            "external_ids": Jsonb(item.external_ids),
            "meta": Jsonb(item.meta),
            "popularity": item.popularity,
        },
    )
    for t in item.translations:
        conn.execute(
            UPSERT_TRANSLATION,
            (item_id, t.locale, t.title, t.synopsis, t.synopsis_license, t.synopsis_source_url),
        )
    for s in item.scores:
        conn.execute(UPSERT_SCORE, (item_id, s.source, s.value, s.url))
    conn.execute(
        "DELETE FROM external_scores WHERE item_id = %s AND source::text = ANY(%s) AND NOT source::text = ANY(%s)",
        (item_id, SCORE_SOURCES, [s.source for s in item.scores]),
    )
    return item_id


def pending_qids(conn: psycopg.Connection) -> list[str]:
    rows = conn.execute(
        "SELECT DISTINCT external_id FROM source_records WHERE normalized_at IS NULL OR normalized_at < fetched_at"
    ).fetchall()
    return sorted(r[0] for r in rows)


def normalize_batch(conn: psycopg.Connection, qids: list[str]) -> tuple[int, int]:
    """Writes one batch inside the caller's transaction; `normalize` commits once per batch."""
    rows = conn.execute(
        "SELECT source, external_id, payload FROM source_records WHERE source = ANY(%s) AND external_id = ANY(%s)",
        (SOURCES, qids),
    ).fetchall()
    by_qid: dict[str, dict[str, dict]] = {}
    for source, qid, payload in rows:
        by_qid.setdefault(qid, {})[source] = payload
    known = known_items(conn, qids)
    saved = skipped = 0
    # Pipeline mode sends the batch's ~8 statements per item without waiting for each reply; one round trip per
    # statement capped a remote database at ~1 item/s. Only a new item's slug check waits for a result.
    with busy(conn), conn.pipeline():
        for qid in qids:
            try:
                item = build_item(qid, by_qid.get(qid, {}))
            except ValidationError as exc:
                log.warning("skipping %s: invalid payload: %s", qid, exc)
                item = None
            item_id = save_item(conn, item, known) if item else None
            saved += item is not None
            skipped += item is None
            conn.execute(
                """UPDATE source_records SET normalized_at = now(), item_id = %s
                WHERE source = ANY(%s) AND external_id = %s""",
                (item_id, SOURCES, qid),
            )
    return saved, skipped


def normalize(conn: psycopg.Connection) -> int:
    qids = pending_qids(conn)
    saved = skipped = 0
    for i in range(0, len(qids), BATCH):
        s, k = normalize_batch(conn, qids[i : i + BATCH])
        conn.commit()
        saved, skipped = saved + s, skipped + k
        log.info("normalized %d/%d (saved=%d skipped=%d)", min(i + BATCH, len(qids)), len(qids), saved, skipped)
    return saved


def main() -> None:
    setup_logging()
    with connect() as conn:
        saved = normalize(conn)
    if saved:
        revalidate(["catalog"])


if __name__ == "__main__":
    main()
