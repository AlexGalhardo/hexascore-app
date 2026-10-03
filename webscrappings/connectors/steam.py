"""Steam `appdetails` public endpoint: header image and OS platforms for games with a Steam app ID."""

import logging
from collections.abc import Iterator
from datetime import datetime

import httpx
import psycopg
from pydantic import BaseModel, ValidationError

from common.cli import run_enrichment
from common.db import Record, existing_ids
from common.http import HttpClient

log = logging.getLogger(__name__)

URL = "https://store.steampowered.com/api/appdetails"
PLATFORM_NAMES = {"windows": "Windows", "mac": "macOS", "linux": "Linux"}


class SteamPlatforms(BaseModel):
    windows: bool = False
    mac: bool = False
    linux: bool = False


class SteamData(BaseModel):
    name: str
    header_image: str | None = None
    platforms: SteamPlatforms = SteamPlatforms()


class SteamEnvelope(BaseModel):
    success: bool
    data: SteamData | None = None


def parse_appdetails(raw: object, appid: str, qid: str) -> dict | None:
    if not isinstance(raw, dict) or appid not in raw:
        return None
    envelope = SteamEnvelope.model_validate(raw[appid])
    if not envelope.success or envelope.data is None:
        return None
    data = envelope.data
    platforms = [name for key, name in PLATFORM_NAMES.items() if getattr(data.platforms, key)]
    return {"qid": qid, "appid": appid, "name": data.name, "header_image": data.header_image, "platforms": platforms}


def targets(conn: psycopg.Connection, since: datetime | None) -> list[tuple[str, str]]:
    return conn.execute(
        """SELECT wikidata_qid, external_ids->>'steam' FROM items
        WHERE type = 'game' AND external_ids ? 'steam'
        AND (%(since)s::timestamptz IS NULL OR updated_at >= %(since)s)
        ORDER BY wikidata_qid""",
        {"since": since},
    ).fetchall()


def fetch(
    since: datetime | None, conn: psycopg.Connection, client: HttpClient | None = None, *, only_missing: bool = False
) -> Iterator[Record]:
    client = client or HttpClient()
    skip = existing_ids(conn, "steam") if only_missing else set()
    for qid, appid in targets(conn, since):
        if not appid.isdigit() or qid in skip:
            continue
        try:
            payload = parse_appdetails(
                client.get_json(URL, params={"appids": appid, "l": "english", "cc": "us"}), appid, qid
            )
        except (httpx.HTTPError, ValidationError) as exc:
            log.warning("steam %s failed: %s", qid, exc)
            continue
        if payload:
            yield Record("steam", qid, payload)


if __name__ == "__main__":
    run_enrichment("steam", fetch)
