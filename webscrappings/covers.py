"""Sets items.meta.coverUrl for items without one.

Source priority: Open Library (books), Steam header then GOG (games), Wikimedia Commons P3383/P18 (all,
with license + author in meta.coverAttribution). With BLOB_READ_WRITE_TOKEN the image is copied to Vercel
Blob; otherwise the source URL is stored as is.
"""

import argparse
import html
import logging
import os
import re
from dataclasses import dataclass
from typing import Any
from urllib.parse import urlsplit

import httpx
import psycopg
from psycopg.types.json import Jsonb
from pydantic import BaseModel, ValidationError

from common.cli import setup_logging
from common.db import connect
from common.http import HttpClient

log = logging.getLogger("covers")

COMMONS_API = "https://commons.wikimedia.org/w/api.php"
BLOB_API = "https://vercel.com/api/blob/"
BLOB_API_VERSION = "12"


@dataclass(frozen=True)
class Cover:
    url: str
    attribution: dict[str, str] | None = None


class MetaValue(BaseModel):
    value: str


class ImageInfo(BaseModel):
    url: str
    thumburl: str | None = None
    descriptionurl: str
    extmetadata: dict[str, MetaValue] = {}


class CommonsPage(BaseModel):
    imageinfo: list[ImageInfo] = []


class CommonsQuery(BaseModel):
    pages: list[CommonsPage]


class CommonsResponse(BaseModel):
    query: CommonsQuery


def commons_file(records: dict[str, dict]) -> str | None:
    claims = records.get("wikidata", {}).get("claims", {})
    for prop in ("P3383", "P18"):
        for claim in claims.get(prop, []):
            if isinstance(claim.get("value"), str):
                return claim["value"]
    return None


def pick_source_cover(item_type: str, records: dict[str, dict]) -> Cover | None:
    """Covers whose URL comes straight from a source record (no extra request needed)."""
    if item_type == "book" and (covers := records.get("openlibrary", {}).get("covers")):
        return Cover(f"https://covers.openlibrary.org/b/id/{covers[0]}-L.jpg")
    if item_type == "game":
        if header := records.get("steam", {}).get("header_image"):
            return Cover(header)
        if gog := records.get("gog", {}).get("cover"):
            return Cover(gog)
    return None


def parse_commons(raw: object) -> Cover | None:
    pages = CommonsResponse.model_validate(raw).query.pages
    if not pages or not pages[0].imageinfo:
        return None
    info = pages[0].imageinfo[0]
    meta = info.extmetadata
    author = html.unescape(re.sub(r"<[^>]+>", "", meta["Artist"].value)).strip() if "Artist" in meta else ""
    license_name = meta["LicenseShortName"].value if "LicenseShortName" in meta else ""
    attribution = {"source": "Wikimedia Commons", "license": license_name, "author": author, "url": info.descriptionurl}
    return Cover(info.thumburl or info.url, attribution)


def commons_cover(client: HttpClient, filename: str) -> Cover | None:
    raw = client.get_json(
        COMMONS_API,
        params={
            "action": "query",
            "titles": f"File:{filename}",
            "prop": "imageinfo",
            "iiprop": "url|extmetadata",
            "iiextmetadatafilter": "Artist|LicenseShortName",
            "iiurlwidth": "600",
            "format": "json",
            "formatversion": "2",
        },
    )
    return parse_commons(raw)


def upload_to_blob(client: HttpClient, token: str, item_id: str, source_url: str) -> str:
    image = client.request("GET", source_url)
    image.raise_for_status()
    content_type = image.headers.get("content-type", "image/jpeg").split(";")[0]
    ext = os.path.splitext(urlsplit(source_url).path)[1].lower() or ".jpg"
    response = client.request(
        "PUT",
        BLOB_API,
        params={"pathname": f"covers/{item_id}{ext}"},
        content=image.content,
        headers={
            "authorization": f"Bearer {token}",
            "x-api-version": BLOB_API_VERSION,
            "x-vercel-blob-access": "public",
            "x-content-type": content_type,
            "x-add-random-suffix": "0",
            "x-allow-overwrite": "1",
        },
    )
    response.raise_for_status()
    return str(response.json()["url"])


def targets(conn: psycopg.Connection, refresh: bool, limit: int | None) -> list[tuple[Any, str, dict[str, dict]]]:
    rows = conn.execute(
        """SELECT i.id, i.type::text, coalesce(jsonb_object_agg(s.source, s.payload)
            FILTER (WHERE s.source IS NOT NULL), '{}'::jsonb)
        FROM items i LEFT JOIN source_records s ON s.item_id = i.id
        WHERE %(refresh)s OR NOT i.meta ? 'coverUrl'
        GROUP BY i.id ORDER BY max(i.popularity) DESC LIMIT %(limit)s""",
        {"refresh": refresh, "limit": limit},
    ).fetchall()
    return [(r[0], r[1], r[2]) for r in rows]


def run(conn: psycopg.Connection, refresh: bool = False, limit: int | None = None) -> int:
    client = HttpClient()
    token = os.environ.get("BLOB_READ_WRITE_TOKEN")
    updated = 0
    for item_id, item_type, records in targets(conn, refresh, limit):
        try:
            cover = pick_source_cover(item_type, records)
            if cover is None and (filename := commons_file(records)):
                cover = commons_cover(client, filename)
            if cover is None:
                continue
            url = upload_to_blob(client, token, str(item_id), cover.url) if token else cover.url
        except (httpx.HTTPError, ValidationError) as exc:
            log.warning("cover for %s failed: %s", item_id, exc)
            continue
        patch: dict[str, Any] = {"coverUrl": url}
        if cover.attribution:
            patch["coverAttribution"] = cover.attribution
        conn.execute("UPDATE items SET meta = (meta - 'coverAttribution') || %s WHERE id = %s", (Jsonb(patch), item_id))
        conn.commit()
        updated += 1
    return updated


def main() -> None:
    setup_logging()
    p = argparse.ArgumentParser(description="Fill items.meta.coverUrl")
    p.add_argument("--refresh", action="store_true", help="recompute covers that are already set")
    p.add_argument("--limit", type=int)
    args = p.parse_args()
    with connect() as conn:
        log.info("covers updated=%d", run(conn, args.refresh, args.limit))


if __name__ == "__main__":
    main()
