"""Sets items.meta.coverUrl (and coverThumbUrl) for items without one.

Source priority: Open Library (books), Steam header then GOG (games), Wikimedia Commons P3383/P18 (all,
with license + author in meta.coverAttribution). With BLOB_READ_WRITE_TOKEN the image is resized, re-encoded
as WebP (full + list thumbnail) and uploaded to Vercel Blob under content-addressed pathnames; otherwise the source
URL is stored as is and coverThumbUrl stays unset.
"""

import argparse
import hashlib
import html
import io
import logging
import os
import re
from concurrent.futures import ThreadPoolExecutor
from dataclasses import dataclass
from itertools import islice
from typing import Any
from urllib.parse import urlsplit

import httpx
import psycopg
from PIL import Image, ImageOps
from psycopg.types.json import Jsonb
from pydantic import BaseModel, ValidationError

from common.cli import setup_logging
from common.db import busy, connect
from common.http import HttpClient

log = logging.getLogger("covers")

COMMONS_API = "https://commons.wikimedia.org/w/api.php"
BLOB_API = "https://vercel.com/api/blob/"
BLOB_API_VERSION = "12"
# The detail page renders covers 176 CSS px tall; 600 covers 3x DPR. Landscape art (Steam headers) is capped at
# 600 wide. thumbnail() never upscales.
MAX_SIZE = (600, 600)
WEBP_QUALITY = 78
# Lists render covers 48 CSS px tall (32x48 / 44x64); 96x144 covers 2x DPR. Same never-upscale fit.
THUMB_SIZE = (96, 144)
# Items processed concurrently (download + encode + upload are latency-bound).
WORKERS = 8
THUMB_QUALITY = 70
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


def to_webp(data: bytes, max_size: tuple[int, int] = MAX_SIZE, quality: int = WEBP_QUALITY) -> bytes:
    """Fits the image into max_size and encodes a WebP without EXIF/XMP/ICC metadata."""
    with Image.open(io.BytesIO(data)) as source:
        image = ImageOps.exif_transpose(source)
        # ponytail: non-sRGB ICC profiles are dropped, not converted, so CMYK/wide-gamut sources (rare for covers)
        # may shift slightly. Upgrade path: PIL.ImageCms.profileToProfile to sRGB before encoding.
        has_alpha = image.mode in ("RGBA", "LA", "PA") or (image.mode == "P" and "transparency" in image.info)
        image = image.convert("RGBA" if has_alpha else "RGB")
        image.thumbnail(max_size, Image.Resampling.LANCZOS)
        out = io.BytesIO()
        image.save(out, "WEBP", quality=quality, method=6)
        return out.getvalue()


def download(client: HttpClient, url: str) -> bytes:
    response = client.request("GET", url)
    response.raise_for_status()
    if len(response.content) > MAX_SOURCE_BYTES:
        raise ValueError(f"source image too large ({len(response.content)} bytes)")
    return response.content


def blob_pathname(item_id: str, source: bytes, suffix: str = "") -> str:
    return f"covers/{item_id}-{hashlib.sha256(source).hexdigest()[:16]}{suffix}.webp"


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


def store_cover(
    client: HttpClient, token: str, item_id: str, source_url: str, current_url: str | None, current_thumb: str | None
) -> tuple[str, str]:
    """Blob URLs of the full and thumbnail WebP copies. Each size is encoded and uploaded only when its stored URL
    does not already point at this source's hash, so items stored before thumbnails existed get just the thumb."""
    source = download(client, source_url)
    # ponytail: when a source image changes, the previous blobs are left orphaned (a few KB each, rare).
    # Upgrade path: delete them through the Blob API delete endpoint after the DB update.
    sizes = ((current_url, "", MAX_SIZE, WEBP_QUALITY), (current_thumb, "-t", THUMB_SIZE, THUMB_QUALITY))
    urls = []
    for current, suffix, max_size, quality in sizes:
        pathname = blob_pathname(item_id, source, suffix)
        if current and urlsplit(current).path.lstrip("/") == pathname:
            urls.append(current)
        else:
            urls.append(upload_to_blob(client, token, pathname, to_webp(source, max_size, quality)))
    return urls[0], urls[1]


def targets(
    conn: psycopg.Connection, refresh: bool, limit: int | None, thumbs: bool = False
) -> list[tuple[Any, str, dict[str, dict], str | None, str | None]]:
    """thumbs (a Blob token is set) also selects covers stored before thumbnails existed or without a token."""
    rows = conn.execute(
        """SELECT i.id, i.type::text, coalesce(jsonb_object_agg(s.source, s.payload)
            FILTER (WHERE s.source IS NOT NULL), '{}'::jsonb), i.meta->>'coverUrl', i.meta->>'coverThumbUrl'
        FROM items i LEFT JOIN source_records s ON s.item_id = i.id
        WHERE %(refresh)s OR NOT i.meta ? 'coverUrl' OR (%(thumbs)s AND NOT i.meta ? 'coverThumbUrl')
        GROUP BY i.id ORDER BY max(i.popularity) DESC LIMIT %(limit)s""",
        {"refresh": refresh, "thumbs": thumbs, "limit": limit},
    ).fetchall()
    return [(r[0], r[1], r[2], r[3], r[4]) for r in rows]


def run(
    conn: psycopg.Connection,
    refresh: bool = False,
    limit: int | None = None,
    client: HttpClient | None = None,
    workers: int = WORKERS,
) -> int:
    client = client or HttpClient()
    token = os.environ.get("BLOB_READ_WRITE_TOKEN")
    if not token:
        log.warning("BLOB_READ_WRITE_TOKEN not set: storing source cover URLs, no WebP copies")

    def patch_for(row: tuple[Any, str, dict[str, dict], str | None, str | None]) -> tuple[Jsonb, Any] | None:
        item_id, item_type, records, current_url, current_thumb = row
        try:
            cover = pick_source_cover(item_type, records)
            if cover is None and (filename := commons_file(records)):
                cover = commons_cover(client, filename)
            if cover is None:
                return None
            url, thumb = (
                store_cover(client, token, str(item_id), cover.url, current_url, current_thumb)
                if token
                else (cover.url, None)
            )
        except (httpx.HTTPError, ValidationError, OSError, ValueError, Image.DecompressionBombError) as exc:
            # OSError covers Pillow's UnidentifiedImageError and truncated files.
            log.warning("cover for %s failed: %s", item_id, exc)
            return None
        except Exception:
            # A malformed stored payload must not abort pool.map and drop the chunk's other patches.
            log.exception("cover for %s failed unexpectedly", item_id)
            return None
        if url == current_url and thumb == current_thumb:
            return None
        patch: dict[str, Any] = {"coverUrl": url}
        if thumb:
            patch["coverThumbUrl"] = thumb
        if cover.attribution:
            patch["coverAttribution"] = cover.attribution
        return Jsonb(patch), item_id

    # Downloads, WebP encodes and Blob uploads run in a thread pool (the HttpClient throttles per host across threads);
    # the DB connection stays on this thread and each chunk is written in one pipelined executemany + commit.
    updated = 0
    pending = iter(targets(conn, refresh, limit, bool(token)))
    with ThreadPoolExecutor(max_workers=workers) as pool:
        while chunk := list(islice(pending, workers * 8)):
            patches = [p for p in pool.map(patch_for, chunk) if p]
            if patches:
                with busy(conn), conn.cursor() as cur:
                    cur.executemany(
                        "UPDATE items SET meta = (meta - 'coverAttribution' - 'coverThumbUrl') || %s WHERE id = %s",
                        patches,
                    )
                conn.commit()
                updated += len(patches)
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
