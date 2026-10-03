import { countItems } from "@/lib/catalog/cached";
import { APP_URL } from "@/lib/config";
import { sitemapIds } from "@/lib/seo";

export const dynamic = "force-dynamic";

// Sitemap index for Search Console: one URL to submit, pointing at every /sitemap/<id>.xml chunk.
export async function GET(): Promise<Response> {
	const count = await countItems().catch((error: unknown) => {
		console.error("[sitemap] item count unavailable, indexing the static sitemap only", error);
		return 0;
	});
	const entries = sitemapIds(count)
		.map((id) => `<sitemap><loc>${APP_URL}/sitemap/${id}.xml</loc></sitemap>`)
		.join("");
	return new Response(
		`<?xml version="1.0" encoding="UTF-8"?><sitemapindex xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">${entries}</sitemapindex>`,
		{ headers: { "content-type": "application/xml", "cache-control": "public, max-age=3600" } },
	);
}
