import { beforeAll, describe, expect, mock, test } from "bun:test";
import { sql } from "drizzle-orm";
import { z } from "zod";
import type { Database } from "@/lib/db/client";
import { items } from "@/lib/db/schema/catalog";
import { createTestDatabase } from "@/lib/db/testing";

// GET /api/items/[id]/ratings against an in-memory PGlite. unstable_cache needs Next's incremental cache, so it is
// replaced by a pass-through; everything else in next/cache stays real.

type Route = (request: Request, context: { params: Promise<{ id: string }> }) => Promise<Response>;
let GET: Route;
let db: Database;

const busy = "01900000-0000-7000-8000-000000000001";
const quiet = "01900000-0000-7000-8000-000000000002";
const unknown = "01900000-0000-7000-8000-000000000099";

const bodySchema = z.object({
	data: z.array(
		z.object({
			nickname: z.string(),
			overall: z.number(),
			categories: z.array(z.number()).length(6),
			ratedAt: z.string(),
		}),
	),
	page: z.number(),
	page_count: z.number(),
});

function call(id: string, query = ""): Promise<Response> {
	return GET(new Request(`http://localhost/api/items/${id}/ratings${query}`), { params: Promise.resolve({ id }) });
}

beforeAll(async () => {
	({ db } = await createTestDatabase());
	const nextCache = await import("next/cache");
	mock.module("next/cache", () => ({ ...nextCache, unstable_cache: <T>(fn: T): T => fn }));
	mock.module("@/lib/db/client", () => ({ db }));
	GET = (await import("@/app/api/items/[id]/ratings/route")).GET;

	await db.insert(items).values([
		{ id: busy, type: "game", slug: "busy", wikidataQid: "Q1" },
		{ id: quiet, type: "game", slug: "quiet", wikidataQid: "Q2" },
	]);
	// Set-based inserts keep the setup fast. Every 10th profile is private: 1135 - 113 = 1022 public ratings on the
	// busy item (past the 1000 browse cap), with pairs sharing a timestamp to exercise the tie-breaker.
	await db.execute(sql`
		insert into "user" (id, name, email, username, profile_private)
		select 'u' || n, 'User ' || n, 'u' || n || '@example.test', 'user' || n, n % 10 = 0
		from generate_series(1, 1135) as n`);
	await db.execute(sql`
		insert into ratings (user_id, item_id, c1, c2, c3, c4, c5, c6, created_at)
		select 'u' || n, ${busy}::uuid, 50, 50, 50, 50, 50, 50, timestamptz '2026-01-01' + (n / 2) * interval '1 second'
		from generate_series(1, 1135) as n`);
	await db.execute(sql`
		insert into ratings (user_id, item_id, c1, c2, c3, c4, c5, c6)
		select 'u' || n, ${quiet}::uuid, 70, 70, 70, 70, 70, 70 from generate_series(1, 12) as n`);
});

describe("GET /api/items/[id]/ratings", () => {
	test("defaults to page 1 with 10 ratings, newest first, cacheable", async () => {
		const response = await call(quiet);
		expect(response.status).toBe(200);
		expect(response.headers.get("cache-control")).toBe("public, max-age=60");
		const body = bodySchema.parse(await response.json());
		// u10 is private: 11 public ratings → 2 pages
		expect([body.page, body.page_count, body.data.length]).toEqual([1, 2, 10]);
		const second = bodySchema.parse(await (await call(quiet, "?page=2")).json());
		expect(second.data).toHaveLength(1);
		expect(new Set([...body.data, ...second.data].map((r) => r.nickname)).size).toBe(11);
	});

	test("caps page_count at 100 and keeps the last page reachable", async () => {
		const first = bodySchema.parse(await (await call(busy)).json());
		expect(first.page_count).toBe(100);
		expect(first.data[0]?.nickname).toBe("user1135");
		const last = bodySchema.parse(await (await call(busy, "?page=100")).json());
		expect([last.page, last.data.length]).toEqual([100, 10]);
		const times = last.data.map((r) => r.ratedAt);
		expect(times).toEqual([...times].sort().reverse());
	});

	test("rejects out-of-range or malformed pages with 400 and unknown items with 404", async () => {
		for (const query of ["?page=0", "?page=101", "?page=1.5", "?page=abc", "?page="]) {
			const response = await call(quiet, query);
			expect(response.status).toBe(400);
			expect(response.headers.get("content-type")).toContain("application/problem+json");
		}
		expect((await call(unknown)).status).toBe(404);
		expect((await call("not-a-uuid")).status).toBe(404);
	});
});
