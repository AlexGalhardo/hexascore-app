"""Sets items.meta.coverUrl for items without one.

Source priority: Open Library (books), Steam header then GOG (games), Wikimedia Commons P3383/P18 (all,
with license + author in meta.coverAttribution). With BLOB_READ_WRITE_TOKEN the image is resized, re-encoded
as WebP and uploaded to Vercel Blob under a content-addressed pathname; otherwise the source URL is stored as is.
"""

import argparse
import hashlib
import html
import io
import logging
import os
import re
from dataclasses import dataclass
from typing import Any
from urllib.parse import urlsplit

import httpx
import psycopg
from PIL import Image, ImageOps
from psycopg.types.json import Jsonb
from pydantic import BaseModel, ValidationError

from common.cli import setup_logging
from common.db import connect
from common.http import HttpClient

log = logging.getLogger("covers")

COMMONS_API = "https://commons.wikimedia.org/w/api.php"
BLOB_API = "https://vercel.com/api/blob/"
BLOB_API_VERSION = "12"
# The detail page renders covers 176 CSS px tall; 600 covers 3x DPR. Landscape art (Steam headers) is capped at
# 600 wide. thumbnail() never upscales.
MAX_SIZE = (600, 600)
WEBP_QUALITY = 78
MAX_SOURCE_BYTES = 15 * 1024 * 1024
# Pathnames are content-addressed, so a blob never changes and can be cached for a year.
BLOB_CACHE_SECONDS = 31_536_000


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


class BlobPut(BaseModel):
    url: str


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


def to_webp(data: bytes) -> bytes:
    """Fits the image into MAX_SIZE and encodes a WebP without EXIF/XMP/ICC metadata."""
    with Image.open(io.BytesIO(data)) as source:
        image = ImageOps.exif_transpose(source)
        # ponytail: non-sRGB ICC profiles are dropped, not converted, so CMYK/wide-gamut sources (rare for covers)
        # may shift slightly. Upgrade path: PIL.ImageCms.profileToProfile to sRGB before encoding.
        has_alpha = image.mode in ("RGBA", "LA", "PA") or (image.mode == "P" and "transparency" in image.info)
        image = image.convert("RGBA" if has_alpha else "RGB")
        image.thumbnail(MAX_SIZE, Image.Resampling.LANCZOS)
        out = io.BytesIO()
        image.save(out, "WEBP", quality=WEBP_QUALITY, method=6)
        return out.getvalue()


def download(client: HttpClient, url: str) -> bytes:
    response = client.request("GET", url)
    response.raise_for_status()
    if len(response.content) > MAX_SOURCE_BYTES:
        raise ValueError(f"source image too large ({len(response.content)} bytes)")
    return response.content


def blob_pathname(item_id: str, source: bytes) -> str:
    return f"covers/{item_id}-{hashlib.sha256(source).hexdigest()[:16]}.webp"


def upload_to_blob(client: HttpClient, token: str, pathname: str, webp: bytes) -> str:
    response = client.request(
        "PUT",
        BLOB_API,
        params={"pathname": pathname},
        content=webp,
        headers={
            "authorization": f"Bearer {token}",
            "x-api-version": BLOB_API_VERSION,
            "x-vercel-blob-access": "public",
            "x-content-type": "image/webp",
            "x-cache-control-max-age": str(BLOB_CACHE_SECONDS),
            "x-add-random-suffix": "0",
            # Same pathname means same source bytes, so overwriting (a retry after a lost response) is harmless.
            "x-allow-overwrite": "1",
        },
    )
    response.raise_for_status()
    return BlobPut.model_validate(response.json()).url


def store_cover(client: HttpClient, token: str, item_id: str, source_url: str, current_url: str | None) -> str:
    """Blob URL of the WebP copy; encode + upload are skipped when the stored blob already has this source's hash."""
    source = download(client, source_url)
    pathname = blob_pathname(item_id, source)
    if current_url and urlsplit(current_url).path.lstrip("/") == pathname:
        return current_url
    # ponytail: when a source image changes, the previous blob is left orphaned (a few KB each, rare).
    # Upgrade path: delete it through the Blob API delete endpoint after the DB update.
    return upload_to_blob(client, token, pathname, to_webp(source))


def targets(
    conn: psycopg.Connection, refresh: bool, limit: int | None
) -> list[tuple[Any, str, dict[str, dict], str | None]]:
    rows = conn.execute(
        """SELECT i.id, i.type::text, coalesce(jsonb_object_agg(s.source, s.payload)
            FILTER (WHERE s.source IS NOT NULL), '{}'::jsonb), i.meta->>'coverUrl'
        FROM items i LEFT JOIN source_records s ON s.item_id = i.id
        WHERE %(refresh)s OR NOT i.meta ? 'coverUrl'
        GROUP BY i.id ORDER BY max(i.popularity) DESC LIMIT %(limit)s""",
        {"refresh": refresh, "limit": limit},
    ).fetchall()
    return [(r[0], r[1], r[2], r[3]) for r in rows]


def run(
    conn: psycopg.Connection, refresh: bool = False, limit: int | None = None, client: HttpClient | None = None
) -> int:
    client = client or HttpClient()
    token = os.environ.get("BLOB_READ_WRITE_TOKEN")
    if not token:
        log.warning("BLOB_READ_WRITE_TOKEN not set: storing source cover URLs, no WebP copies")
    updated = 0
    for item_id, item_type, records, current_url in targets(conn, refresh, limit):
        try:
            cover = pick_source_cover(item_type, records)
            if cover is None and (filename := commons_file(records)):
                cover = commons_cover(client, filename)
            if cover is None:
                continue
            url = store_cover(client, token, str(item_id), cover.url, current_url) if token else cover.url
        except (httpx.HTTPError, ValidationError, OSError, ValueError, Image.DecompressionBombError) as exc:
            # OSError covers Pillow's UnidentifiedImageError and truncated files.
            log.warning("cover for %s failed: %s", item_id, exc)
            continue
        if url == current_url:
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
