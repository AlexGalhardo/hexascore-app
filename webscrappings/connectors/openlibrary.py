"""Open Library API connector: page count, cover ids and description for books with an ISBN-13 or Open Library ID."""

import logging
import re
from collections.abc import Iterator
from datetime import datetime

import httpx
import psycopg
from pydantic import BaseModel, ValidationError

from common.cli import run_enrichment
from common.db import Record
from common.http import HttpClient

log = logging.getLogger(__name__)

BASE = "https://openlibrary.org"


class Ref(BaseModel):
    key: str


class Text(BaseModel):
    value: str


class OpenLibraryDoc(BaseModel):
    key: str
    number_of_pages: int | None = None
    covers: list[int] = []
    works: list[Ref] = []
    description: str | Text | None = None

    @property
    def cover_ids(self) -> list[int]:
        # Open Library uses -1 as a "no cover" placeholder.
        return [c for c in self.covers if c > 0]

    @property
    def overview(self) -> str | None:
        text = self.description.value if isinstance(self.description, Text) else self.description
        return text.strip() or None if text else None


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


def work_path(raw: object) -> str | None:
    """Path of the work an edition belongs to; None for a work document or an unexpected key."""
    works = OpenLibraryDoc.model_validate(raw).works
    return f"{works[0].key}.json" if works and re.fullmatch(r"/works/OL\d+W", works[0].key) else None


def parse_doc(raw: object, qid: str, work: object | None = None) -> dict:
    """`work` (the edition's work document) fills the cover and the description the edition lacks. The description
    is stored only as input for the synopsis rewriting step; it is never copied into `item_translations`."""
    doc = OpenLibraryDoc.model_validate(raw)
    parent = OpenLibraryDoc.model_validate(work) if work else None
    pages = doc.number_of_pages if doc.number_of_pages and doc.number_of_pages > 0 else None
    covers = doc.cover_ids or (parent.cover_ids if parent else [])
    overview = doc.overview or (parent.overview if parent else None)
    return {"qid": qid, "key": doc.key, "pages": pages, "covers": covers[:3], "overview": overview}


def targets(conn: psycopg.Connection, since: datetime | None) -> list[tuple[str, dict[str, str], bool]]:
    """The flag says the item has an English Wikipedia synopsis, so no description is needed for it."""
    return conn.execute(
        """SELECT wikidata_qid, external_ids, EXISTS (SELECT 1 FROM source_records w
            WHERE w.source = 'wikipedia:en' AND w.external_id = items.wikidata_qid) FROM items
        WHERE type = 'book' AND (external_ids ? 'isbn13' OR external_ids ? 'openlibrary')
        AND (%(since)s::timestamptz IS NULL OR updated_at >= %(since)s)
        ORDER BY wikidata_qid""",
        {"since": since},
    ).fetchall()


def complete_ids(conn: psycopg.Connection) -> set[str]:
    """Records a resume can skip: fetched in the current shape (`overview` key), or older ones with nothing left
    to fill (a cover and a Wikipedia synopsis)."""
    rows = conn.execute(
        """SELECT o.external_id FROM source_records o WHERE o.source = 'openlibrary' AND (o.payload ? 'overview'
            OR (o.payload #>> '{covers,0}' IS NOT NULL AND EXISTS (SELECT 1 FROM source_records w
                WHERE w.source = 'wikipedia:en' AND w.external_id = o.external_id)))"""
    )
    return {r[0] for r in rows}


def fetch(
    since: datetime | None, conn: psycopg.Connection, client: HttpClient | None = None, *, only_missing: bool = False
) -> Iterator[Record]:
    client = client or HttpClient()
    skip = complete_ids(conn) if only_missing else set()

    def get(path: str) -> object | None:
        response = client.request("GET", BASE + path)
        if response.status_code == httpx.codes.NOT_FOUND:
            return None
        response.raise_for_status()
        return response.json()

    for qid, external_ids, has_synopsis in targets(conn, since):
        path = document_path(external_ids)
        if not path or qid in skip:
            continue
        try:
            raw = get(path)
            if raw is None:
                continue
            doc = OpenLibraryDoc.model_validate(raw)
            # The second request is only spent when the edition leaves something to fill.
            needs_work = not doc.cover_ids or not (has_synopsis or doc.overview)
            parent = get(parent_path) if needs_work and (parent_path := work_path(raw)) else None
            payload = parse_doc(raw, qid, parent)
        except (httpx.HTTPError, ValidationError) as exc:
            log.warning("openlibrary %s failed: %s", qid, exc)
            continue
        yield Record("openlibrary", qid, payload)


if __name__ == "__main__":
    run_enrichment("openlibrary", fetch)
