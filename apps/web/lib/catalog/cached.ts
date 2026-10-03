import { unstable_cache } from "next/cache";
import { cache } from "react";
import type { Locale } from "@/i18n/routing";
import { db } from "@/lib/db/client";
import type { ItemType } from "@/lib/db/schema/catalog";
import * as repo from "./repository";
import type { AffiliateLink, ExternalScore, ItemDetail, ListItemsParams, PublicRatingsPage } from "./types";

// Next 16 without `cacheComponents`: `"use cache"` + cacheTag is unavailable, so catalog reads go through
// unstable_cache (Data Cache, survives deploys) tagged `catalog` and `item:<id>`. The ingestion invalidates them
// via POST /api/cron/revalidate. Results must be JSON-serializable — the repository DTOs already are.

export const CATALOG_TAG = "catalog";
export const itemTag = (id: string): string => `item:${id}`;
const HOUR = 3600;

export const listItems = unstable_cache((params: ListItemsParams) => repo.listItems(db, params), ["catalog:list"], {
	tags: [CATALOG_TAG],
	revalidate: HOUR,
});

export const countListItems = unstable_cache(
	(params: ListItemsParams) => repo.countListItems(db, params),
	["catalog:list-count"],
	{ tags: [CATALOG_TAG], revalidate: HOUR },
);

export const searchItems = unstable_cache(
	(q: string, options: { type?: ItemType; locale: Locale; limit?: number }) => repo.searchItems(db, q, options),
	["catalog:search"],
	{ tags: [CATALOG_TAG], revalidate: HOUR },
);

const DAY = 24 * HOUR;

export const countItems = unstable_cache(() => repo.countItems(db), ["catalog:count"], {
	tags: [CATALOG_TAG],
	revalidate: DAY,
});

export const listSitemapItems = unstable_cache(
	(offset: number, limit: number) => repo.listSitemapItems(db, offset, limit),
	["catalog:sitemap"],
	{ tags: [CATALOG_TAG], revalidate: DAY },
);

export const listGenres = unstable_cache((type: ItemType) => repo.listGenres(db, type), ["catalog:genres"], {
	tags: [CATALOG_TAG],
	revalidate: HOUR,
});

export const listPlatforms = unstable_cache((type: ItemType) => repo.listPlatforms(db, type), ["catalog:platforms"], {
	tags: [CATALOG_TAG],
	revalidate: HOUR,
});

const itemIdBySlug = unstable_cache(
	(type: ItemType, slug: string) => repo.getItemIdBySlug(db, type, slug),
	["catalog:slug"],
	{ tags: [CATALOG_TAG], revalidate: HOUR },
);

// Item-scoped entries carry `item:<id>`, so a rating write (0.3) or a re-ingested item refreshes only that page.
function forItem<T>(id: string, key: string, load: () => Promise<T>): Promise<T> {
	return unstable_cache(load, [`catalog:${key}`, id], { tags: [CATALOG_TAG, itemTag(id)], revalidate: HOUR })();
}

// React cache() dedupes generateMetadata + page within one request.
export const getItemBySlug = cache(async (type: ItemType, slug: string, locale: Locale): Promise<ItemDetail | null> => {
	const id = await itemIdBySlug(type, slug);
	if (!id) return null;
	const item = await forItem(id, `item:${locale}`, () => repo.getItemById(db, id, locale));
	return item?.type === type ? item : null;
});

export function getExternalScores(id: string): Promise<ExternalScore[]> {
	return forItem(id, "external", () => repo.getExternalScores(db, id));
}

export function getAffiliateLinks(id: string): Promise<AffiliateLink[]> {
	return forItem(id, "affiliates", () => repo.getAffiliateLinks(db, id));
}

/** `id` must be a validated UUID and `page` an integer in 1..RATINGS_MAX_PAGES. */
export function getItemRatingsPage(id: string, page: number): Promise<PublicRatingsPage> {
	return forItem(id, `ratings:p${page}`, () => repo.getItemRatingsPage(db, id, page));
}

export function itemExists(id: string): Promise<boolean> {
	return forItem(id, "exists", () => repo.itemExists(db, id));
}
