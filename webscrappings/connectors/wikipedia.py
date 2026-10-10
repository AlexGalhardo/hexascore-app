"""Wikipedia intro extracts per locale (CC BY-SA 4.0), via the Action API in batches of titles.

Targets are the en/pt/es sitelinks of Wikidata records fetched since `since`. The REST `page/summary`
endpoint answers one page per request and is rate limited hard for anonymous clients (a few req/s per
IP across all wikis); `action=query&prop=extracts` returns up to 20 intros per request.
"""

import logging
from collections.abc import Iterable, Iterator
from concurrent.futures import ThreadPoolExecutor
from datetime import datetime
from itertools import batched, islice

import httpx
import psycopg
from pydantic import BaseModel, ValidationError

from common.cli import parser, setup_logging
from common.db import Record, connect, existing_ids, store_records
from common.http import HttpClient

log = logging.getLogger(__name__)

LANGS = ("en", "pt", "es")
# Concurrent batch requests per run; the per-host throttle in HttpClient still caps the rate.
WORKERS = 4
# TextExtracts caps `exintro` results at 20 pages per request.
BATCH = 20


class Page(BaseModel):
    title: str
    extract: str = ""
    fullurl: str = ""
    pageprops: dict[str, str] = {}


class Query(BaseModel):
    pages: list[Page] = []


class QueryResponse(BaseModel):
    query: Query = Query()


def parse_extracts(raw: object, qids: set[str], lang: str) -> list[dict]:
    payloads = []
    for page in QueryResponse.model_validate(raw).query.pages:
        qid = page.pageprops.get("wikibase_item")
        # A redirect can land on a different entity; a disambiguation page has no usable synopsis.
        if qid not in qids or "disambiguation" in page.pageprops or not page.extract.strip():
            continue
        payloads.append(
            {"qid": qid, "lang": lang, "title": page.title, "extract": page.extract.strip(), "url": page.fullurl}
        )
    return payloads


def targets(conn: psycopg.Connection, since: datetime | None) -> list[tuple[str, dict[str, str]]]:
    return conn.execute(
        """SELECT external_id, payload->'sitelinks' FROM source_records
        WHERE source = 'wikidata' AND (%(since)s::timestamptz IS NULL OR fetched_at >= %(since)s)
        ORDER BY external_id""",
        {"since": since},
    ).fetchall()


def fetch(
    since: datetime | None,
    conn: psycopg.Connection,
    client: HttpClient | None = None,
    *,
    only_missing: bool = False,
    langs: Iterable[str] = LANGS,
    workers: int = WORKERS,
) -> Iterator[Record]:
    client = client or HttpClient()
    langs = [lang for lang in langs if lang in LANGS]
    skip = {lang: existing_ids(conn, f"wikipedia:{lang}") if only_missing else set() for lang in langs}

    rows = targets(conn, since)

    def tasks() -> Iterator[tuple[str, list[tuple[str, str]]]]:
        for lang in langs:
            pairs = (
                (qid, sitelinks[f"{lang}wiki"])
                for qid, sitelinks in rows
                if sitelinks.get(f"{lang}wiki") and qid not in skip[lang]
            )
            for batch in batched(pairs, BATCH, strict=False):
                yield lang, list(batch)

    def extracts(task: tuple[str, list[tuple[str, str]]]) -> list[Record]:
        lang, batch = task
        params = {
            "action": "query",
            "format": "json",
            "formatversion": "2",
            "redirects": "1",
            "prop": "extracts|pageprops|info",
            "exintro": "1",
            "explaintext": "1",
            "exlimit": str(BATCH),
            "ppprop": "wikibase_item|disambiguation",
            "inprop": "url",
            "titles": "|".join(title for _, title in batch),
        }
        try:
            response = client.request("GET", f"https://{lang}.wikipedia.org/w/api.php", params=params)
            response.raise_for_status()
            payloads = parse_extracts(response.json(), {qid for qid, _ in batch}, lang)
        except (httpx.HTTPError, ValidationError) as exc:
            log.warning("wikipedia %s batch of %d from %s failed: %s", lang, len(batch), batch[0][0], exc)
            return []
        return [Record(f"wikipedia:{lang}", p["qid"], p) for p in payloads]

    # Requests are latency-bound, so a small pool reaches the per-host rate limit; the shared client throttles
    # per host. Work is submitted in bounded chunks so memory stays flat, and records are yielded (and stored)
    # on the caller's thread, which owns the DB connection.
    pending = tasks()
    with ThreadPoolExecutor(max_workers=workers) as pool:
        while chunk := list(islice(pending, workers * 8)):
            for records in pool.map(extracts, chunk):
                yield from records


def main() -> None:
    setup_logging()
    p = parser("Fetch Wikipedia summaries into source_records")
    p.add_argument("--langs", default=",".join(LANGS), help="comma-separated Wikipedia languages")
    args = p.parse_args()
    with connect() as conn:
        records = fetch(args.since, conn, only_missing=args.missing_only, langs=args.langs.split(","))
        seen, changed = store_records(conn, records)
    log.info("wikipedia records fetched=%d changed=%d", seen, changed)


if __name__ == "__main__":
    main()
