"""GOG public catalog API: availability, store link and vertical cover for games with a GOG ID on Wikidata.

Wikidata stores the store path (`game/<slug>`), which the catalog can't filter on, so we search by
title and pick the product whose slug matches.
"""

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

URL = "https://catalog.gog.com/v1/catalog"
OS_NAMES = {"windows": "Windows", "osx": "macOS", "linux": "Linux"}


class GogProduct(BaseModel):
    id: str
    slug: str
    title: str
    storeLink: str
    coverVertical: str | None = None
    operatingSystems: list[str] = []
    productState: str | None = None


class GogCatalog(BaseModel):
    products: list[GogProduct]


def _norm(text: str) -> str:
    return re.sub(r"[^a-z0-9]+", "", text.lower())


def pick_product(raw: object, gog_id: str, title: str) -> GogProduct | None:
    products = GogCatalog.model_validate(raw).products
    slug = gog_id.removeprefix("game/").removeprefix("movie/")
    for match in (
        lambda p: p.slug == slug,
        lambda p: _norm(p.title) == _norm(title),
        lambda p: p.slug.startswith(slug),
    ):
        if found := next((p for p in products if match(p)), None):
            return found
    return None


def parse_catalog(raw: object, gog_id: str, title: str, qid: str) -> dict | None:
    product = pick_product(raw, gog_id, title)
    if product is None:
        return None
    return {
        "qid": qid,
        "id": product.id,
        "slug": product.slug,
        "title": product.title,
        "store_url": product.storeLink,
        "cover": product.coverVertical,
        "platforms": [OS_NAMES[o] for o in product.operatingSystems if o in OS_NAMES],
        "state": product.productState,
    }


def targets(conn: psycopg.Connection, since: datetime | None) -> list[tuple[str, str, str]]:
    return conn.execute(
        """SELECT i.wikidata_qid, i.external_ids->>'gog', t.title FROM items i
        JOIN item_translations t ON t.item_id = i.id AND t.locale = 'en'
        WHERE i.type = 'game' AND i.external_ids ? 'gog'
        AND (%(since)s::timestamptz IS NULL OR i.updated_at >= %(since)s)
        ORDER BY i.wikidata_qid""",
        {"since": since},
    ).fetchall()


def fetch(
    since: datetime | None, conn: psycopg.Connection, client: HttpClient | None = None, *, only_missing: bool = False
) -> Iterator[Record]:
    client = client or HttpClient()
    skip = existing_ids(conn, "gog") if only_missing else set()
    for qid, gog_id, title in targets(conn, since):
        if qid in skip:
            continue
        params = {
            "limit": 10,
            "query": f"like:{title}",
            "productType": "in:game,pack",
            "countryCode": "US",
            "locale": "en-US",
            "currencyCode": "USD",
        }
        try:
            payload = parse_catalog(client.get_json(URL, params=params), gog_id, title, qid)
        except (httpx.HTTPError, ValidationError) as exc:
            log.warning("gog %s failed: %s", qid, exc)
            continue
        if payload:
            yield Record("gog", qid, payload)


if __name__ == "__main__":
    run_enrichment("gog", fetch)
