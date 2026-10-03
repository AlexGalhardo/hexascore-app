import { z } from "zod";
import { getItemRatingsPage, itemExists } from "@/lib/catalog/cached";
import { RATINGS_MAX_PAGES } from "@/lib/catalog/repository";
import { problem } from "@/lib/problem";

const paramsSchema = z.object({ id: z.uuid() });
const querySchema = z.object({ page: z.coerce.number().int().min(1).max(RATINGS_MAX_PAGES).default(1) });

// Public ratings of one item (public profiles only), used by the "See users' scores" dialog. Numbered pages of 10
// over the newest 1000 ratings: page_count = min(100, ceil(min(total, 1000) / 10)).
export async function GET(request: Request, { params }: { params: Promise<{ id: string }> }): Promise<Response> {
	const parsedParams = paramsSchema.safeParse(await params);
	if (!parsedParams.success) return problem(404, "Not Found", "Unknown item.");
	const query = querySchema.safeParse(Object.fromEntries(new URL(request.url).searchParams));
	if (!query.success)
		return problem(400, "Bad Request", `page must be an integer between 1 and ${RATINGS_MAX_PAGES}.`);

	const { id } = parsedParams.data;
	try {
		if (!(await itemExists(id))) return problem(404, "Not Found", "Unknown item.");
		const result = await getItemRatingsPage(id, query.data.page);
		return Response.json(
			{ data: result.data, page: result.page, page_count: result.pageCount },
			{ headers: { "cache-control": "public, max-age=60" } },
		);
	} catch (error) {
		console.error("[api] item ratings failed", { id, error });
		return problem(500, "Internal Server Error");
	}
}
