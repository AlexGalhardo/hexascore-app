import { apiRoute } from "@/lib/api/handler";
import { searchQuerySchema, searchResultSchema, toItemSummary } from "@/lib/api/schemas";
import { searchItems } from "@/lib/catalog/repository";
import { db } from "@/lib/db/client";

export const GET = apiRoute(
	{ query: (limits) => searchQuerySchema(limits.maxPageSize), response: searchResultSchema },
	async (_ctx, query) => {
		const results = await searchItems(db, query.q, { type: query.type, locale: query.locale, limit: query.limit });
		return { data: results.map((item) => toItemSummary(item, query.locale)) };
	},
);
