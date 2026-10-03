import { z } from "zod";
import { routing } from "@/i18n/routing";
import { searchItems } from "@/lib/catalog/cached";
import { ITEM_TYPES } from "@/lib/db/schema/catalog";
import { problem } from "@/lib/problem";
import { itemHref } from "@/lib/sections";

const querySchema = z.object({
	q: z.string().trim().max(100).default(""),
	locale: z.enum(routing.locales).default(routing.defaultLocale),
	type: z.enum(ITEM_TYPES).optional(),
});

// Navbar autocomplete. Returns locale-less paths; the client prefixes the locale when navigating.
export async function GET(request: Request): Promise<Response> {
	const query = querySchema.safeParse(Object.fromEntries(new URL(request.url).searchParams));
	if (!query.success) return problem(400, "Bad Request", "Invalid query parameters.");
	const { q, locale, type } = query.data;
	if (q.length < 2) return Response.json({ data: [] });

	try {
		const results = await searchItems(q.toLowerCase(), { locale, type, limit: 8 });
		return Response.json(
			{
				data: results.map((r) => ({
					id: r.id,
					type: r.type,
					title: r.title,
					year: r.year,
					score: r.score,
					coverUrl: r.coverUrl,
					coverThumbUrl: r.coverThumbUrl,
					path: itemHref(r.type, r.slug),
				})),
			},
			{ headers: { "cache-control": "public, max-age=60" } },
		);
	} catch (error) {
		console.error("[api] search suggest failed", { q, error });
		return problem(500, "Internal Server Error");
	}
}
