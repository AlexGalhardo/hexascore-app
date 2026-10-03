import { revalidateTag } from "next/cache";
import { z } from "zod";
import { isCronAuthorized } from "@/lib/cron-auth";
import { env } from "@/lib/env";
import { problem } from "@/lib/problem";

// Tags used by lib/catalog/cached.ts: `catalog` (everything) and `item:<uuid>` (one item page).
const bodySchema = z.object({
	tags: z
		.array(z.string().regex(/^(catalog|item:[0-9a-f-]{36})$/, "must be `catalog` or `item:<uuid>`"))
		.min(1)
		.max(1000),
});

// Called by the Python ingestion after writing to the catalog.
export async function POST(request: Request): Promise<Response> {
	if (!env.CRON_SECRET) return problem(503, "Service Unavailable", "CRON_SECRET is not configured.");
	if (!isCronAuthorized(request.headers.get("authorization"), env.CRON_SECRET)) {
		return problem(401, "Unauthorized", undefined, { "www-authenticate": "Bearer" });
	}

	let json: unknown;
	try {
		json = await request.json();
	} catch {
		return problem(400, "Bad Request", "Body must be JSON.");
	}
	const body = bodySchema.safeParse(json);
	if (!body.success) return problem(400, "Bad Request", z.prettifyError(body.error));

	const tags = [...new Set(body.data.tags)];
	// "max": stale-while-revalidate — visitors keep getting the old page while the next request refreshes it.
	for (const tag of tags) revalidateTag(tag, "max");
	console.info(`[cron] revalidated ${tags.length} tag(s)`);
	return Response.json({ revalidated: tags });
}
