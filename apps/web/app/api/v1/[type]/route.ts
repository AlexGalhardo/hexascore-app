import { apiRoute } from "@/lib/api/handler";
import { itemListSchema, listQuerySchema, toItemSummary, typeParamsSchema } from "@/lib/api/schemas";
import { listItems } from "@/lib/catalog/repository";
import { CATEGORY_COLUMNS, type CategoryColumn } from "@/lib/catalog/types";
import { CATEGORY_KEYS } from "@/lib/categories";
import { db } from "@/lib/db/client";

// Catalog reads hit the repository directly (indexed queries, `private, no-store` responses); add a Data Cache
// layer only if the k6 p95 budget in 0.6 requires it.
export const GET = apiRoute(
	{
		params: typeParamsSchema,
		query: (limits, { type }) => listQuerySchema(type, limits.maxPageSize),
		response: itemListSchema,
	},
	async (_ctx, query, { type }) => {
		const values = query as Record<string, unknown>;
		const minCategory: Partial<Record<CategoryColumn, number>> = {};
		CATEGORY_KEYS[type].forEach((key, i) => {
			const min = values[`min_${key}`];
			const column = CATEGORY_COLUMNS[i];
			if (typeof min === "number" && column) minCategory[column] = min;
		});
		const page = await listItems(db, {
			type,
			locale: query.locale,
			genre: query.genre,
			platform: typeof values.platform === "string" ? values.platform : undefined,
			year: query.year,
			minScore: query.min_score,
			maxScore: query.max_score,
			minCategory,
			sort: query.sort,
			limit: query.limit,
			cursor: query.cursor,
		});
		return { data: page.data.map((item) => toItemSummary(item, query.locale)), next_cursor: page.nextCursor };
	},
);
