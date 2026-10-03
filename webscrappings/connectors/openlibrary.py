"""Open Library API connector: page count and cover ids for books with an ISBN-13 or Open Library ID."""

import logging
import re
from collections.abc import Iterator
from datetime import datetime

import httpx
import psycopg
from pydantic import BaseModel, ValidationError

from common.cli import run_enrichment
from common.db import Record, existing_ids
from common.http import HttpClient

log = logging.getLogger(__name__)

BASE = "https://openlibrary.org"


class OpenLibraryDoc(BaseModel):
    key: str
    number_of_pages: int | None = None
    covers: list[int] = []


def document_path(external_ids: dict[str, str]) -> str | None:
    """Edition data (ISBN or OL...M) carries the page count, so it wins over a work (OL...W)."""
    isbn = re.sub(r"[^0-9X]", "", external_ids.get("isbn13", "").upper())
    olid = external_ids.get("openlibrary", "")
    if len(isbn) == 13:
        return f"/isbn/{isbn}.json"
    if re.fullmatch(r"OL\d+M", olid):
        return f"/books/{olid}.json"
    if re.fullmatch(r"OL\d+W", olid):
        return f"/works/{olid}.json"
    return None


def parse_doc(raw: object, qid: str) -> dict:
    doc = OpenLibraryDoc.model_validate(raw)
    pages = doc.number_of_pages if doc.number_of_pages and doc.number_of_pages > 0 else None
    # Open Library uses -1 as a "no cover" placeholder.
    return {"qid": qid, "key": doc.key, "pages": pages, "covers": [c for c in doc.covers if c > 0][:3]}


def targets(conn: psycopg.Connection, since: datetime | None) -> list[tuple[str, dict[str, str]]]:
    return conn.execute(
        """SELECT wikidata_qid, external_ids FROM items
        WHERE type = 'book' AND (external_ids ? 'isbn13' OR external_ids ? 'openlibrary')
        AND (%(since)s::timestamptz IS NULL OR updated_at >= %(since)s)
        ORDER BY wikidata_qid""",
        {"since": since},
    ).fetchall()


def fetch(
    since: datetime | None, conn: psycopg.Connection, client: HttpClient | None = None, *, only_missing: bool = False
) -> Iterator[Record]:
    client = client or HttpClient()
    skip = existing_ids(conn, "openlibrary") if only_missing else set()
    for qid, external_ids in targets(conn, since):
        path = document_path(external_ids)
        if not path or qid in skip:
            continue
        try:
            response = client.request("GET", BASE + path)
            if response.status_code == httpx.codes.NOT_FOUND:
                continue
            response.raise_for_status()
            payload = parse_doc(response.json(), qid)
        except (httpx.HTTPError, ValidationError) as exc:
            log.warning("openlibrary %s failed: %s", qid, exc)
            continue
        yield Record("openlibrary", qid, payload)


if __name__ == "__main__":
    run_enrichment("openlibrary", fetch)
