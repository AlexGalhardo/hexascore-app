"""Wikidata connector: SPARQL selects QIDs, `wbgetentities` returns the same entity JSON as the dump.

The stored payload is a trimmed entity (only the properties the normalizer reads, plus English labels
for referenced items), so unrelated edits on Wikidata don't change the hash.
"""

import argparse
import logging
from collections.abc import Iterable, Iterator
from datetime import datetime
from typing import Literal

from pydantic import BaseModel, JsonValue

from common.cli import parse_since, setup_logging
from common.db import Record, connect, store_records
from common.http import HttpClient

log = logging.getLogger(__name__)

SPARQL_URL = "https://query.wikidata.org/sparql"
API_URL = "https://www.wikidata.org/w/api.php"
BATCH = 50

ItemType = Literal["game", "movie", "series", "book"]
TYPES: tuple[ItemType, ...] = ("movie", "series", "game", "book")
TYPE_CLASSES: dict[ItemType, tuple[str, ...]] = {
    "movie": ("Q11424", "Q24869", "Q202866"),
    "series": ("Q5398426", "Q581714", "Q1259759", "Q526877"),
    "game": ("Q7889",),
    "book": ("Q7725634", "Q571", "Q8261"),
}
BOOTSTRAP_TARGETS: dict[ItemType, int] = {"movie": 35_000, "series": 15_000, "game": 25_000, "book": 25_000}
INCREMENTAL_MIN_SITELINKS = 15

# "mul" holds labels valid in every language (common for names); used as a fallback.
LABEL_LANGS = ("en", "mul", "pt-br", "pt", "es")
SITELINKS = ("enwiki", "ptwiki", "eswiki")
REF_PROPS = ("P136", "P400", "P57", "P170", "P178", "P50", "P495")
SCALAR_PROPS = (
    "P577", "P580", "P571",  # dates
    "P2047", "P1104", "P2437", "P1113",  # runtime, pages, seasons, episodes
    "P1651", "P18", "P3383", "P444",  # trailers, images, review scores
    "P345", "P5794", "P1733", "P4947", "P4983", "P212", "P648", "P2816",
    "P8383", "P1712", "P1258", "P6278", "P2725",
)  # fmt: skip
KEPT_PROPS = frozenset(REF_PROPS + SCALAR_PROPS)
SCORE_QUALIFIERS = ("P447", "P459", "P585")


class RawDataValue(BaseModel):
    type: str
    value: JsonValue


class RawSnak(BaseModel):
    snaktype: str
    datavalue: RawDataValue | None = None


class RawStatement(BaseModel):
    mainsnak: RawSnak
    qualifiers: dict[str, list[RawSnak]] = {}
    rank: Literal["preferred", "normal", "deprecated"]


class RawTerm(BaseModel):
    value: str


class RawSitelink(BaseModel):
    title: str


class RawEntity(BaseModel):
    id: str
    missing: str | None = None
    labels: dict[str, RawTerm] = {}
    claims: dict[str, list[RawStatement]] = {}
    sitelinks: dict[str, RawSitelink] = {}


class Claim(BaseModel):
    value: JsonValue
    rank: Literal["preferred", "normal"]
    qualifiers: dict[str, list[JsonValue]] = {}


class WikidataPayload(BaseModel):
    """Trimmed entity stored in source_records.payload for source='wikidata'."""

    type: ItemType
    id: str
    labels: dict[str, str]
    sitelinks: dict[str, str]
    sitelinks_count: int
    claims: dict[str, list[Claim]]
    refs: dict[str, str] = {}


def simplify_value(dv: RawDataValue | None) -> JsonValue:
    if dv is None:
        return None
    v = dv.value
    match dv.type:
        case "wikibase-entityid" if isinstance(v, dict):
            return v.get("id")
        case "string":
            return v
        case "monolingualtext" if isinstance(v, dict):
            return v.get("text")
        case "time" if isinstance(v, dict):
            return {"time": v.get("time"), "precision": v.get("precision")}
        case "quantity" if isinstance(v, dict):
            unit = str(v.get("unit", "1"))
            return {"amount": v.get("amount"), "unit": unit.rsplit("/", 1)[-1]}
    return None


def trim_entity(raw: RawEntity, item_type: ItemType) -> dict:
    claims: dict[str, list[dict]] = {}
    for prop, statements in raw.claims.items():
        if prop not in KEPT_PROPS:
            continue
        kept = []
        for st in statements:
            value = simplify_value(st.mainsnak.datavalue)
            if st.rank == "deprecated" or value is None:
                continue
            claim: dict = {"value": value, "rank": st.rank}
            if prop == "P444":
                claim["qualifiers"] = {
                    q: [simplify_value(s.datavalue) for s in st.qualifiers[q]]
                    for q in SCORE_QUALIFIERS
                    if q in st.qualifiers
                }
            kept.append(claim)
        if kept:
            claims[prop] = kept
    return {
        "type": item_type,
        "id": raw.id,
        "labels": {lang: raw.labels[lang].value for lang in LABEL_LANGS if lang in raw.labels},
        "sitelinks": {site: raw.sitelinks[site].title for site in SITELINKS if site in raw.sitelinks},
        "sitelinks_count": len(raw.sitelinks),
        "claims": claims,
    }


def referenced_qids(payload: dict) -> set[str]:
    return {c["value"] for prop in REF_PROPS for c in payload["claims"].get(prop, []) if isinstance(c["value"], str)}


def sparql(client: HttpClient, query: str) -> list[dict[str, dict[str, str]]]:
    data = client.get_json(
        SPARQL_URL,
        params={"query": query, "format": "json"},
        headers={"Accept": "application/sparql-results+json"},
    )
    return data["results"]["bindings"]


def _qid(uri: str) -> str:
    return uri.rsplit("/", 1)[-1]


def _values(item_type: ItemType) -> str:
    return " ".join(f"wd:{c}" for c in TYPE_CLASSES[item_type])


def top_query(item_type: ItemType, limit: int) -> str:
    return f"""SELECT DISTINCT ?item ?links WHERE {{
  VALUES ?class {{ {_values(item_type)} }}
  ?item wdt:P31 ?class ; wikibase:sitelinks ?links .
}} ORDER BY DESC(?links) LIMIT {int(limit)}"""


def modified_query(item_type: ItemType, since: datetime, min_sitelinks: int) -> str:
    ts = since.strftime("%Y-%m-%dT%H:%M:%SZ")
    return f"""SELECT DISTINCT ?item WHERE {{
  VALUES ?class {{ {_values(item_type)} }}
  ?item wdt:P31 ?class ; wikibase:sitelinks ?links ; schema:dateModified ?mod .
  FILTER(?links >= {int(min_sitelinks)} && ?mod >= "{ts}"^^xsd:dateTime)
}}"""


def select_qids(client: HttpClient, item_type: ItemType, since: datetime | None, limit: int | None) -> list[str]:
    if since is None:
        query = top_query(item_type, limit or BOOTSTRAP_TARGETS[item_type])
    else:
        query = modified_query(item_type, since, INCREMENTAL_MIN_SITELINKS)
        if limit:
            query += f" LIMIT {int(limit)}"
    return [_qid(row["item"]["value"]) for row in sparql(client, query)]


def get_entities(client: HttpClient, ids: list[str], props: str, languages: Iterable[str]) -> list[RawEntity]:
    data = client.get_json(
        API_URL,
        params={
            "action": "wbgetentities",
            "ids": "|".join(ids),
            "props": props,
            "languages": "|".join(languages),
            "format": "json",
        },
    )
    return [RawEntity.model_validate(e) for e in data.get("entities", {}).values()]


class LabelCache:
    def __init__(self, client: HttpClient) -> None:
        self._client = client
        self._labels: dict[str, str] = {}

    def resolve(self, qids: set[str]) -> dict[str, str]:
        missing = sorted(q for q in qids if q not in self._labels)
        for i in range(0, len(missing), BATCH):
            chunk = missing[i : i + BATCH]
            for e in get_entities(self._client, chunk, "labels", ["en", "mul"]):
                label = e.labels.get("en") or e.labels.get("mul")
                self._labels[e.id] = label.value if label else ""
        return {q: self._labels[q] for q in qids if self._labels.get(q)}


def fetch_entities(client: HttpClient, typed_qids: Iterable[tuple[str, ItemType]]) -> Iterator[Record]:
    labels = LabelCache(client)
    pending = list(dict.fromkeys(typed_qids))
    types = dict(pending)
    for i in range(0, len(pending), BATCH):
        chunk = [q for q, _ in pending[i : i + BATCH]]
        entities = get_entities(client, chunk, "labels|claims|sitelinks", LABEL_LANGS)
        for raw in entities:
            if raw.missing is not None or raw.id not in types:
                continue
            payload = trim_entity(raw, types[raw.id])
            payload["refs"] = labels.resolve(referenced_qids(payload))
            yield Record("wikidata", raw.id, WikidataPayload.model_validate(payload).model_dump())


def fetch(
    since: datetime | None,
    *,
    types: Iterable[ItemType] = TYPES,
    limit: int | None = None,
    client: HttpClient | None = None,
) -> Iterator[Record]:
    """since=None selects the top items by sitelinks (bootstrap); otherwise items edited since then."""
    client = client or HttpClient()
    for item_type in types:
        qids = select_qids(client, item_type, since, limit)
        log.info("wikidata %s: %d qids selected", item_type, len(qids))
        yield from fetch_entities(client, ((q, item_type) for q in qids))


def known_items(conn) -> list[tuple[str, ItemType]]:
    return conn.execute("SELECT wikidata_qid, type::text FROM items ORDER BY wikidata_qid").fetchall()


def main() -> None:
    setup_logging()
    p = argparse.ArgumentParser(description="Fetch Wikidata entities into source_records")
    p.add_argument("--since", type=parse_since, help="items edited since (36h, 8d, ISO); omit for top-by-sitelinks")
    p.add_argument("--types", default=",".join(TYPES), help="comma-separated item types")
    p.add_argument("--limit", type=int, help="max items per type")
    p.add_argument("--refresh-known", action="store_true", help="re-fetch every item already in the catalog")
    args = p.parse_args()
    types = [t for t in args.types.split(",") if t in TYPES]
    with connect() as conn:
        if args.refresh_known:
            records = fetch_entities(HttpClient(), known_items(conn))
        else:
            records = fetch(args.since, types=types, limit=args.limit)
        seen, changed = store_records(conn, records)
    log.info("wikidata records fetched=%d changed=%d", seen, changed)


if __name__ == "__main__":
    main()
