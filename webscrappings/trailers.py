"""Checks which candidate trailers can still be embedded and records, per item, the ones `normalize` may show.

Candidates are the YouTube ids from Wikidata (P1651) and from the TMDB / IGDB records, in that order. Each one is
asked to YouTube's oEmbed endpoint: 200 means public and embeddable; 400/401/403/404 mean removed, private or with
embedding disabled (the "Video unavailable" player). An item stops being checked once MAX_TRAILERS ids are available.
The result is a `youtube` source record: {"available": [...], "unavailable": [...]}.
"""

import logging
import re
from collections.abc import Iterator
from concurrent.futures import ThreadPoolExecutor
from datetime import datetime
from itertools import islice
from typing import Any

import httpx
import psycopg
from pydantic import BaseModel, ValidationError

from common.cli import run_enrichment
from common.db import Record
from common.http import HttpClient
from connectors.wikidata import Claim

log = logging.getLogger("trailers")

OEMBED = "https://www.youtube.com/oembed"
MAX_TRAILERS = 2
WORKERS = 16
YOUTUBE_ID = re.compile(r"[A-Za-z0-9_-]{11}")
UNAVAILABLE = {400, 401, 403, 404}


class Checked(BaseModel):
    available: list[str] = []
    unavailable: list[str] = []


def candidates(records: dict[str, Any]) -> list[str]:
    """Well-formed YouTube ids in display order: Wikidata (preferred rank first), then TMDB / IGDB."""
    claims = [Claim.model_validate(c) for c in records.get("wikidata") or []]
    ids = [str(c.value) for c in sorted(claims, key=lambda c: c.rank != "preferred")]
    for source in ("tmdb", "igdb"):
        ids += [str(v) for v in records.get(source) or []]
    return list(dict.fromkeys(i for i in ids if YOUTUBE_ID.fullmatch(i)))


def is_available(client: HttpClient, youtube_id: str) -> bool:
    response = client.request(
        "GET", OEMBED, params={"url": f"https://www.youtube.com/watch?v={youtube_id}", "format": "json"}
    )
    if response.status_code in UNAVAILABLE:
        return False
    # Anything else that is not a success (429 or 5xx after the retries) says nothing about the video.
    response.raise_for_status()
    return True


def check(client: HttpClient, ids: list[str], known: Checked) -> Checked:
    """`known` holds earlier verdicts to reuse; pass an empty one to ask YouTube about every id again."""
    result = Checked()
    for youtube_id in ids:
        if len(result.available) == MAX_TRAILERS:
            break
        if youtube_id in known.available:
            ok = True
        elif youtube_id in known.unavailable:
            ok = False
        else:
            ok = is_available(client, youtube_id)
        (result.available if ok else result.unavailable).append(youtube_id)
    return result


def targets(conn: psycopg.Connection, since: datetime | None) -> list[tuple[str, dict[str, Any]]]:
    # Only the trailer keys of each payload, like covers.targets: whole Wikidata entities are gigabytes.
    return conn.execute(
        """SELECT external_id, jsonb_object_agg(source, CASE source
                WHEN 'wikidata' THEN payload #> '{claims,P1651}'
                WHEN 'youtube' THEN payload
                ELSE payload -> 'trailers' END)
        FROM source_records WHERE source IN ('wikidata', 'tmdb', 'igdb', 'youtube')
        GROUP BY external_id
        HAVING %(since)s::timestamptz IS NULL OR max(fetched_at) FILTER (WHERE source <> 'youtube') >= %(since)s
        ORDER BY external_id""",
        {"since": since},
    ).fetchall()


def fetch(
    since: datetime | None,
    conn: psycopg.Connection,
    client: HttpClient | None = None,
    *,
    only_missing: bool = False,
    workers: int = WORKERS,
) -> Iterator[Record]:
    """`only_missing` reuses the verdicts already stored and only asks YouTube about ids never checked."""
    client = client or HttpClient()

    def checked(target: tuple[str, dict[str, Any]]) -> Record | None:
        qid, records = target
        try:
            ids = candidates(records)
            stored = records.get("youtube")
            if not ids and stored is None:
                return None
            known = Checked.model_validate(stored) if only_missing and stored else Checked()
            return Record("youtube", qid, {"qid": qid, **check(client, ids, known).model_dump()})
        except (httpx.HTTPError, ValidationError) as exc:
            log.warning("trailers %s failed: %s", qid, exc)
            return None

    # Same shape as the TMDB connector: bounded chunks on a pool, records yielded on the caller's thread.
    pending = iter(targets(conn, since))
    with ThreadPoolExecutor(max_workers=workers) as pool:
        while chunk := list(islice(pending, workers * 8)):
            for record in pool.map(checked, chunk):
                if record:
                    yield record


if __name__ == "__main__":
    run_enrichment("trailers", fetch)
