"""Wikipedia REST `page/summary` connector: synopsis per locale (CC BY-SA 4.0).

Targets are the en/pt/es sitelinks of Wikidata records fetched since `since`.
"""

import logging
from collections.abc import Iterable, Iterator
from concurrent.futures import ThreadPoolExecutor
from datetime import datetime
from itertools import islice
from urllib.parse import quote

import httpx
import psycopg
from pydantic import BaseModel, ValidationError

from common.cli import parser, setup_logging
from common.db import Record, connect, existing_ids, store_records
from common.http import HttpClient

log = logging.getLogger(__name__)

LANGS = ("en", "pt", "es")
# Concurrent summary requests per run; the per-host throttle in HttpClient still caps the rate.
WORKERS = 8


class DesktopUrls(BaseModel):
    page: str


class ContentUrls(BaseModel):
    desktop: DesktopUrls


class Summary(BaseModel):
    type: str
    title: str
    extract: str = ""
    wikibase_item: str | None = None
    content_urls: ContentUrls


def parse_summary(raw: object, qid: str, lang: str) -> dict | None:
    summary = Summary.model_validate(raw)
    # A redirect can land on a different entity; a disambiguation page has no usable synopsis.
    if summary.type != "standard" or summary.wikibase_item != qid or not summary.extract.strip():
        return None
    return {
        "qid": qid,
        "lang": lang,
        "title": summary.title,
        "extract": summary.extract.strip(),
        "url": summary.content_urls.desktop.page,
    }


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

    def tasks() -> Iterator[tuple[str, str, str]]:
        for qid, sitelinks in targets(conn, since):
            for lang in langs:
                title = sitelinks.get(f"{lang}wiki")
                if title and qid not in skip[lang]:
                    yield qid, lang, title

    def summary(task: tuple[str, str, str]) -> Record | None:
        qid, lang, title = task
        url = f"https://{lang}.wikipedia.org/api/rest_v1/page/summary/{quote(title.replace(' ', '_'), safe='')}"
        try:
            response = client.request("GET", url)
            if response.status_code == httpx.codes.NOT_FOUND:
                return None
            response.raise_for_status()
            payload = parse_summary(response.json(), qid, lang)
        except (httpx.HTTPError, ValidationError) as exc:
            log.warning("wikipedia %s %s failed: %s", qid, lang, exc)
            return None
        return Record(f"wikipedia:{lang}", qid, payload) if payload else None

    # Requests are latency-bound, so a small pool reaches the per-host rate limit; the shared client throttles
    # per host. Work is submitted in bounded chunks so memory stays flat, and records are yielded (and stored)
    # on the caller's thread, which owns the DB connection.
    pending = tasks()
    with ThreadPoolExecutor(max_workers=workers) as pool:
        while chunk := list(islice(pending, workers * 8)):
            for record in pool.map(summary, chunk):
                if record:
                    yield record


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
