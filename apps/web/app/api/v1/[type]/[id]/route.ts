import { apiRoute, NotFoundError } from "@/lib/api/handler";
import { detailQuerySchema, itemDetailSchema, itemParamsSchema, toItemDetail } from "@/lib/api/schemas";
import { getItemById } from "@/lib/catalog/repository";
import { db } from "@/lib/db/client";

export const GET = apiRoute(
	{ params: itemParamsSchema, query: detailQuerySchema, response: itemDetailSchema },
	async (_ctx, query, { type, id }) => {
		const item = await getItemById(db, id, query.locale);
		// An id of another type is a 404 too, so `/movies/{game id}` never leaks a game under the wrong path.
		if (!item || item.type !== type) throw new NotFoundError(`No ${type} with id ${id}.`);
		return toItemDetail(item, query.locale);
	},
);
