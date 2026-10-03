import { apiRoute } from "@/lib/api/handler";
import { emptyQuerySchema, genreListSchema, typeParamsSchema } from "@/lib/api/schemas";
import { listGenreCounts } from "@/lib/catalog/repository";
import { db } from "@/lib/db/client";

export const GET = apiRoute(
	{ params: typeParamsSchema, query: emptyQuerySchema, response: genreListSchema },
	async (_ctx, _query, { type }) => {
		const genres = await listGenreCounts(db, type);
		return { data: genres.map((g) => ({ slug: g.value, count: g.count })) };
	},
);
