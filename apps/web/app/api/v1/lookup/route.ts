import { apiRoute, NotFoundError } from "@/lib/api/handler";
import { itemDetailSchema, lookupQuerySchema, toItemDetail } from "@/lib/api/schemas";
import { getItemById, getItemIdByExternalId } from "@/lib/catalog/repository";
import { db } from "@/lib/db/client";

export const GET = apiRoute({ query: lookupQuerySchema, response: itemDetailSchema }, async (_ctx, query) => {
	const id = await getItemIdByExternalId(db, query.source, query.id);
	const item = id ? await getItemById(db, id, query.locale) : null;
	if (!item) throw new NotFoundError(`No item with ${query.source} id ${query.id}.`);
	return toItemDetail(item, query.locale);
});
