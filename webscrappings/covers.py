"""Sets items.meta.coverUrl (and coverThumbUrl) for items without one.

Source priority: Open Library (books), TMDB poster (movies, series), IGDB cover then Steam header then GOG (games),
Wikimedia Commons P3383 film poster (with license + author in meta.coverAttribution). TMDB and IGDB covers carry
their source and page URL in meta.coverAttribution. Commons P18 is not a cover source, and a full run removes the
covers it produced in the past. With BLOB_READ_WRITE_TOKEN the image is resized, re-encoded
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
# The detail page renders covers up to 320 CSS px tall (384 wide for landscape art); 600 covers ~2x DPR.
# Landscape art (Steam headers) is capped at 600 wide. thumbnail() never upscales.
MAX_SIZE = (600, 600)
WEBP_QUALITY = 78
# Lists render covers 48 CSS px tall (32x48 / 44x64); 96x144 covers 2x DPR. Same never-upscale fit.
THUMB_SIZE = (96, 144)
# Items processed concurrently (download + encode + upload are latency-bound).
WORKERS = 32
THUMB_QUALITY = 70
MAX_SOURCE_BYTES = 15 * 1024 * 1024
# Pathnames are content-addressed, so a blob never changes and can be cached for a year.
BLOB_CACHE_SECONDS = 31_536_000
# (item types, source_records.source, meta.coverAttribution.source). The attribution name is also how `targets`
# tells that an item already carries the official cover: it must equal upper(source).
OFFICIAL_SOURCES = ((("movie", "series"), "tmdb", "TMDB"), (("game",), "igdb", "IGDB"))


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
    """Only P3383 (film poster). P18 is any image of the subject (a photo of an arcade cabinet, an author
    portrait), so it is not used as a cover."""
    for claim in records.get("wikidata", {}).get("claims", {}).get("P3383", []):
        if isinstance(claim.get("value"), str):
            return claim["value"]
    return None


def drop_loose_commons_covers(conn: psycopg.Connection) -> int:
    """Removes covers taken from Commons for items without a P3383 poster: those came from P18. Items with an
    official source were already re-covered by this run; the rest stay without a cover."""
    with busy(conn):
        removed = conn.execute(
            """UPDATE items i SET meta = i.meta - 'coverUrl' - 'coverThumbUrl' - 'coverAttribution'
            WHERE i.meta #>> '{coverAttribution,source}' = 'Wikimedia Commons'
            AND NOT EXISTS (SELECT 1 FROM source_records w WHERE w.item_id = i.id AND w.source = 'wikidata'
                AND w.payload #>> '{claims,P3383,0,value}' IS NOT NULL)"""
        ).rowcount
    conn.commit()
    return removed


def pick_source_cover(item_type: str, records: dict[str, dict]) -> Cover | None:
    """Covers whose URL comes straight from a source record (no extra request needed)."""
    if item_type == "book" and (covers := records.get("openlibrary", {}).get("covers")):
        return Cover(f"https://covers.openlibrary.org/b/id/{covers[0]}-L.jpg")
    # The official poster / box art wins over everything else for its type.
    for item_types, source, name in OFFICIAL_SOURCES:
        record = records.get(source, {})
        if item_type in item_types and (url := record.get("cover")):
            return Cover(url, {"source": name} | ({"url": record["url"]} if record.get("url") else {}))
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
    conn: psycopg.Connection, refresh: bool, limit: int | None, thumbs: bool = False, shard: tuple[int, int] = (0, 1)
) -> list[tuple[Any, str, dict[str, dict], str | None, str | None]]:
    """thumbs (a Blob token is set) also selects covers stored before thumbnails existed or without a token.
    shard (index, count) keeps only the items whose id hashes to that index, so several processes can split a run.
    Items whose cover came from a lesser source are selected again once a better one has an image: a TMDB/IGDB
    cover replaces anything not attributed to it, an Open Library cover replaces a Commons image."""
    # Only the sources and keys pick_source_cover/commons_file read: full payloads (whole Wikidata entities,
    # Wikipedia extracts) are gigabytes over 100k items and kept the query streaming for 15+ minutes.
    rows = conn.execute(
        """SELECT i.id, i.type::text, coalesce(jsonb_object_agg(s.source, CASE s.source
                WHEN 'wikidata' THEN jsonb_build_object('claims', jsonb_strip_nulls(jsonb_build_object(
                    'P3383', s.payload #> '{claims,P3383}')))
                ELSE jsonb_strip_nulls(jsonb_build_object('covers', s.payload -> 'covers',
                    'header_image', s.payload -> 'header_image', 'cover', s.payload -> 'cover',
                    'url', s.payload -> 'url'))
            END) FILTER (WHERE s.source IS NOT NULL), '{}'::jsonb), i.meta->>'coverUrl', i.meta->>'coverThumbUrl'
        FROM items i LEFT JOIN source_records s
            ON s.item_id = i.id AND s.source IN ('wikidata', 'openlibrary', 'steam', 'gog', 'tmdb', 'igdb')
        WHERE (hashtext(i.id::text) & 2147483647) %% %(shards)s = %(shard)s
        GROUP BY i.id
        -- The better-source test reads the joined rows (bool_or) on purpose: source_records has no index on
        -- item_id, so the same test as a correlated EXISTS scans the whole table once per item.
        HAVING %(refresh)s OR NOT i.meta ? 'coverUrl' OR (%(thumbs)s AND NOT i.meta ? 'coverThumbUrl')
            OR bool_or(s.source IN ('tmdb', 'igdb') AND s.payload ->> 'cover' IS NOT NULL
                AND i.meta #>> '{coverAttribution,source}' IS DISTINCT FROM upper(s.source))
            OR bool_or(s.source = 'openlibrary' AND s.payload #>> '{covers,0}' IS NOT NULL
                AND i.meta #>> '{coverAttribution,source}' = 'Wikimedia Commons')
        ORDER BY max(i.popularity) DESC LIMIT %(limit)s""",
        {"refresh": refresh, "thumbs": thumbs, "limit": limit, "shard": shard[0], "shards": shard[1]},
    ).fetchall()
    return [(r[0], r[1], r[2], r[3], r[4]) for r in rows]


def run(
    conn: psycopg.Connection,
    refresh: bool = False,
    limit: int | None = None,
    client: HttpClient | None = None,
    workers: int = WORKERS,
    shard: tuple[int, int] = (0, 1),
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
    pending = iter(targets(conn, refresh, limit, bool(token), shard))
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
    # A limited or sharded run has not offered every item its official cover yet, so it must not strip the old ones.
    if limit is None and shard[1] == 1:
        log.info("loose Commons covers removed=%d", drop_loose_commons_covers(conn))
    return updated


def main() -> None:
    setup_logging()
    p = argparse.ArgumentParser(description="Fill items.meta.coverUrl")
    p.add_argument("--refresh", action="store_true", help="recompute covers that are already set")
    p.add_argument("--limit", type=int)
    # WebP encoding is CPU-bound, so one process tops out near one core's worth of covers; shards scale it.
    p.add_argument("--shard", default="0/1", help="INDEX/COUNT: process only this slice of the items (parallel runs)")
    args = p.parse_args()
    index, count = (int(part) for part in args.shard.split("/"))
    if not 0 <= index < count:
        p.error("--shard must be INDEX/COUNT with 0 <= INDEX < COUNT")
    with connect() as conn:
        log.info("covers updated=%d", run(conn, args.refresh, args.limit, shard=(index, count)))


if __name__ == "__main__":
    main()
