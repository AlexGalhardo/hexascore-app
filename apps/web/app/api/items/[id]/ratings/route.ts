import { z } from "zod";
import { getItemRatingsPage, itemExists } from "@/lib/catalog/cached";
import { problem } from "@/lib/problem";

const paramsSchema = z.object({ id: z.uuid() });
const querySchema = z.object({ cursor: z.string().min(1).max(512).optional() });

// Public ratings of one item (public profiles only), used by the "See users' scores" dialog.
export async function GET(request: Request, { params }: { params: Promise<{ id: string }> }): Promise<Response> {
	const parsedParams = paramsSchema.safeParse(await params);
	if (!parsedParams.success) return problem(404, "Not Found", "Unknown item.");
	const query = querySchema.safeParse(Object.fromEntries(new URL(request.url).searchParams));
	if (!query.success) return problem(400, "Bad Request", "Invalid query parameters.");

	const { id } = parsedParams.data;
	try {
		if (!(await itemExists(id))) return problem(404, "Not Found", "Unknown item.");
		const page = await getItemRatingsPage(id, query.data.cursor);
		return Response.json(
			{ data: page.data, next_cursor: page.nextCursor },
			{ headers: { "cache-control": "public, max-age=60" } },
		);
	} catch (error) {
		if (error instanceof Error && error.name === "InvalidCursorError") {
			return problem(400, "Bad Request", "Invalid cursor.");
		}
		console.error("[api] item ratings failed", { id, error });
		return problem(500, "Internal Server Error");
	}
}
