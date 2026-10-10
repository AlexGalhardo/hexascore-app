"""IGDB API connector: official cover and missing facts for games with an IGDB slug (Wikidata P5794).

Authenticates with a Twitch app access token (IGDB_TWITCH_CLIENT_ID + IGDB_TWITCH_CLIENT_SECRET, client
credentials grant). The summary is stored only as input for the synopsis rewriting step; it is never copied into
`item_translations`.
"""

import logging
import os
import re
from collections.abc import Iterator
from datetime import UTC, datetime, timedelta
from itertools import batched

import httpx
import psycopg
from pydantic import BaseModel, ValidationError

from common.cli import run_enrichment
from common.db import Record, existing_ids
from common.http import HttpClient

log = logging.getLogger(__name__)

TOKEN_URL = "https://id.twitch.tv/oauth2/token"
GAMES_URL = "https://api.igdb.com/v4/games"
# 528x748, the largest portrait cover preset.
COVER_URL = "https://images.igdb.com/igdb/image/upload/t_cover_big_2x/{image_id}.jpg"
FIELDS = (
    "slug,summary,first_release_date,cover.image_id,genres.name,platforms.name,"
    "involved_companies.developer,involved_companies.company.name,videos.video_id"
)
# One request returns at most 500 games.
BATCH = 200
MAX_TRAILERS = 3
EPOCH = datetime(1970, 1, 1, tzinfo=UTC)
# Slugs go into the query body, so anything outside IGDB's slug alphabet is dropped instead of escaped.
SLUG = re.compile(r"[a-z0-9][a-z0-9-]*")


class Token(BaseModel):
    access_token: str


class Named(BaseModel):
    name: str


class Cover(BaseModel):
    image_id: str


class InvolvedCompany(BaseModel):
    developer: bool = False
    company: Named | None = None


class Video(BaseModel):
    video_id: str


class Game(BaseModel):
    id: int
    slug: str
    summary: str | None = None
    first_release_date: int | None = None
    cover: Cover | None = None
    genres: list[Named] = []
    platforms: list[Named] = []
    involved_companies: list[InvolvedCompany] = []
    videos: list[Video] = []


def games_query(slugs: list[str]) -> str:
    quoted = ",".join(f'"{s}"' for s in slugs)
    return f"fields {FIELDS}; where slug = ({quoted}); limit {len(slugs)};"


def parse_games(raw: object, qids: dict[str, str]) -> list[dict]:
    """`qids` maps slug -> QID for the requested batch; games outside it are ignored."""
    if not isinstance(raw, list):
        raise ValueError("IGDB response is not a list")
    payloads = []
    for entry in raw:
        game = Game.model_validate(entry)
        if game.slug not in qids:
            continue
        # Epoch arithmetic instead of fromtimestamp, which raises OSError on Windows for pre-1970 releases.
        released = (EPOCH + timedelta(seconds=game.first_release_date)).date() if game.first_release_date else None
        developers = [c.company.name for c in game.involved_companies if c.developer and c.company]
        payloads.append(
            {
                "qid": qids[game.slug],
                "id": game.id,
                "slug": game.slug,
                "url": f"https://www.igdb.com/games/{game.slug}",
                "cover": COVER_URL.format(image_id=game.cover.image_id) if game.cover else None,
                "overview": game.summary or None,
                "release_date": released.isoformat() if released else None,
                "genres": [g.name for g in game.genres],
                "platforms": [p.name for p in game.platforms],
                "creators": list(dict.fromkeys(developers)),
                "trailers": [v.video_id for v in game.videos][:MAX_TRAILERS],
            }
        )
    return payloads


def access_token(client: HttpClient, client_id: str, client_secret: str) -> str:
    response = client.request(
        "POST",
        TOKEN_URL,
        params={"client_id": client_id, "client_secret": client_secret, "grant_type": "client_credentials"},
    )
    response.raise_for_status()
    return Token.model_validate(response.json()).access_token


def targets(conn: psycopg.Connection, since: datetime | None) -> list[tuple[str, str]]:
    return conn.execute(
        """SELECT wikidata_qid, external_ids->>'igdb' FROM items
        WHERE type = 'game' AND external_ids ? 'igdb'
        AND (%(since)s::timestamptz IS NULL OR updated_at >= %(since)s)
        ORDER BY wikidata_qid""",
        {"since": since},
    ).fetchall()


def fetch(
    since: datetime | None, conn: psycopg.Connection, client: HttpClient | None = None, *, only_missing: bool = False
) -> Iterator[Record]:
    client_id = os.environ.get("IGDB_TWITCH_CLIENT_ID")
    client_secret = os.environ.get("IGDB_TWITCH_CLIENT_SECRET")
    if not client_id or not client_secret:
        raise SystemExit("IGDB_TWITCH_CLIENT_ID and IGDB_TWITCH_CLIENT_SECRET must be set")
    client = client or HttpClient()
    skip = existing_ids(conn, "igdb") if only_missing else set()
    pairs = [(qid, slug) for qid, slug in targets(conn, since) if qid not in skip and SLUG.fullmatch(slug)]
    if not pairs:
        return
    headers = {"Client-ID": client_id, "Authorization": f"Bearer {access_token(client, client_id, client_secret)}"}
    for batch in batched(pairs, BATCH, strict=False):
        qids = {slug: qid for qid, slug in batch}
        try:
            response = client.request("POST", GAMES_URL, content=games_query(list(qids)), headers=headers)
            response.raise_for_status()
            payloads = parse_games(response.json(), qids)
        except (httpx.HTTPError, ValidationError, ValueError) as exc:
            log.warning("igdb batch of %d from %s failed: %s", len(batch), batch[0][0], exc)
            continue
        for payload in payloads:
            yield Record("igdb", payload["qid"], payload)


if __name__ == "__main__":
    run_enrichment("igdb", fetch)
