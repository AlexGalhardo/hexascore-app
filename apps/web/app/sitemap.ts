import type { MetadataRoute } from "next";
import { countItems, listSitemapItems } from "@/lib/catalog/cached";
import { APP_URL } from "@/lib/config";
import { itemEntries, SITEMAP_CHUNK, STATIC_PATHS, sitemapEntry, sitemapIds } from "@/lib/seo";

// Rendered per request from the Data Cache (1 day, `catalog` tag): a build-time prerender would need a reachable
// database in every build environment.
export const dynamic = "force-dynamic";

export async function generateSitemaps(): Promise<{ id: number }[]> {
	try {
		return sitemapIds(await countItems()).map((id) => ({ id }));
	} catch (error) {
		console.error("[sitemap] item count unavailable, serving static pages only", error);
		return [{ id: 0 }];
	}
}

export default async function sitemap(props: { id: Promise<string> }): Promise<MetadataRoute.Sitemap> {
	const id = Number(await props.id);
	if (!Number.isInteger(id) || id < 0) return [];
	if (id === 0) return STATIC_PATHS.map((path) => sitemapEntry(APP_URL, path));
	return itemEntries(APP_URL, await listSitemapItems((id - 1) * SITEMAP_CHUNK, SITEMAP_CHUNK));
}
