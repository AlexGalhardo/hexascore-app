"""TMDB API connector: official poster and missing facts for movies and series with a TMDB or IMDb ID.

Uses the v3 endpoints with the v4 read access token (THEMOVIEDATABASE_API_READ_ACCESS_TOKEN) as a Bearer token.
The overview is stored only as input for the synopsis rewriting step; it is never copied into `item_translations`.
"""

import logging
import os
import re
from collections.abc import Iterator
from concurrent.futures import ThreadPoolExecutor
from datetime import datetime
from itertools import islice

import httpx
import psycopg
from pydantic import BaseModel, ValidationError

from common.cli import run_enrichment
from common.db import Record, existing_ids
from common.http import HttpClient

log = logging.getLogger(__name__)

API = "https://api.themoviedb.org/3"
# w780 is the smallest poster size above the 600 px the cover pipeline keeps.
POSTER_BASE = "https://image.tmdb.org/t/p/w780"
MEDIA = {"movie": "movie", "series": "tv"}
WORKERS = 8
MAX_TRAILERS = 3


class Named(BaseModel):
    name: str


class Video(BaseModel):
    key: str
    site: str = ""
    type: str = ""
    official: bool = False


class Videos(BaseModel):
    results: list[Video] = []


class CrewMember(BaseModel):
    name: str
    job: str = ""


class Credits(BaseModel):
    crew: list[CrewMember] = []


class Details(BaseModel):
    id: int
    overview: str | None = None
    poster_path: str | None = None
    runtime: int | None = None
    episode_run_time: list[int] = []
    genres: list[Named] = []
    release_date: str | None = None
    first_air_date: str | None = None
    number_of_seasons: int | None = None
    number_of_episodes: int | None = None
    created_by: list[Named] = []
    production_countries: list[Named] = []
    credits: Credits = Credits()
    videos: Videos = Videos()


class FindResult(BaseModel):
    id: int


class FindResponse(BaseModel):
    movie_results: list[FindResult] = []
    tv_results: list[FindResult] = []


def parse_find(raw: object, media: str) -> int | None:
    found = FindResponse.model_validate(raw)
    results = found.movie_results if media == "movie" else found.tv_results
    return results[0].id if results else None


def parse_details(raw: object, media: str, qid: str) -> dict:
    d = Details.model_validate(raw)
    released = d.release_date if media == "movie" else d.first_air_date
    runtime = d.runtime if media == "movie" else next(iter(d.episode_run_time), None)
    people = [c for c in d.credits.crew if c.job == "Director"] if media == "movie" else d.created_by
    trailers = sorted(
        (v for v in d.videos.results if v.site == "YouTube" and v.type == "Trailer"), key=lambda v: not v.official
    )
    return {
        "qid": qid,
        "id": d.id,
        "url": f"https://www.themoviedb.org/{media}/{d.id}",
        "cover": POSTER_BASE + d.poster_path if d.poster_path else None,
        "overview": d.overview or None,
        "runtime": runtime or None,
        "genres": [g.name for g in d.genres],
        "release_date": released if released and re.fullmatch(r"\d{4}-\d\d-\d\d", released) else None,
        "seasons": d.number_of_seasons or None,
        "episodes": d.number_of_episodes or None,
        "creators": list(dict.fromkeys(p.name for p in people)),
        "countries": [c.name for c in d.production_countries],
        "trailers": list(dict.fromkeys(v.key for v in trailers))[:MAX_TRAILERS],
    }


def targets(conn: psycopg.Connection, since: datetime | None) -> list[tuple[str, str, str | None, str | None]]:
    return conn.execute(
        """SELECT wikidata_qid, type::text, external_ids->>'tmdb', external_ids->>'imdb' FROM items
        WHERE type IN ('movie', 'series') AND (external_ids ? 'tmdb' OR external_ids ? 'imdb')
        AND (%(since)s::timestamptz IS NULL OR updated_at >= %(since)s)
        ORDER BY wikidata_qid""",
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
    token = os.environ.get("THEMOVIEDATABASE_API_READ_ACCESS_TOKEN")
    if not token:
        raise SystemExit("THEMOVIEDATABASE_API_READ_ACCESS_TOKEN is not set")
    client = client or HttpClient()
    headers = {"Authorization": f"Bearer {token}", "Accept": "application/json"}
    skip = existing_ids(conn, "tmdb") if only_missing else set()

    def get(path: str, params: dict[str, str]) -> object | None:
        response = client.request("GET", API + path, params=params, headers=headers)
        if response.status_code == httpx.codes.UNAUTHORIZED:
            # Not an httpx error on purpose: a rejected token must stop the run, not log one warning per item.
            raise PermissionError("TMDB rejected the access token")
        if response.status_code == httpx.codes.NOT_FOUND:
            return None
        response.raise_for_status()
        return response.json()

    def details(target: tuple[str, str, str | None, str | None]) -> Record | None:
        qid, item_type, tmdb_id, imdb_id = target
        media = MEDIA[item_type]
        try:
            if not (tmdb_id and tmdb_id.isdigit()):
                # Wikidata has no TMDB ID: resolve it through the IMDb ID (exact match, never by title).
                if not (imdb_id and re.fullmatch(r"tt\d+", imdb_id)):
                    return None
                found = get(f"/find/{imdb_id}", {"external_source": "imdb_id"})
                resolved = parse_find(found, media) if found else None
                if resolved is None:
                    return None
                tmdb_id = str(resolved)
            append = "credits,videos" if media == "movie" else "videos"
            raw = get(f"/{media}/{tmdb_id}", {"append_to_response": append})
            return Record("tmdb", qid, parse_details(raw, media, qid)) if raw else None
        except (httpx.HTTPError, ValidationError) as exc:
            log.warning("tmdb %s failed: %s", qid, exc)
            return None

    # Same shape as the Wikipedia connector: bounded chunks on a small pool, records yielded on the caller's
    # thread (which owns the DB connection); the shared client throttles the host across threads.
    pending = (t for t in targets(conn, since) if t[0] not in skip)
    with ThreadPoolExecutor(max_workers=workers) as pool:
        while chunk := list(islice(pending, workers * 8)):
            for record in pool.map(details, chunk):
                if record:
                    yield record


if __name__ == "__main__":
    run_enrichment("tmdb", fetch)
