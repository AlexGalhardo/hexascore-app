import { beforeAll, beforeEach, describe, expect, mock, test } from "bun:test";
import { eq } from "drizzle-orm";
import type { z } from "zod";
import type { Database } from "@/lib/db/client";
import { apiKeys, subscriptions, timeToBeat, user } from "@/lib/db/schema";
import { externalScores, items, itemTranslations } from "@/lib/db/schema/catalog";
import { createTestDatabase } from "@/lib/db/testing";
import { generateApiKey } from "./keys";
import { type CounterStore, getStore } from "./quota";
import { itemDetailSchema, itemListSchema, problemSchema, usageSchema } from "./schemas";

// Route-level tests: the exported GET handlers run against an in-memory PGlite (db client mocked) and the
// in-memory quota store, exactly as `next start` would call them.

type Route = (request: Request, context: { params: Promise<unknown> }) => Promise<Response>;
let routes: Record<"list" | "detail" | "genres" | "search" | "lookup" | "usage", Route>;
let db: Database;
let store: CounterStore;
const throttle = { off: true };

const uuid = (n: number) => `01900000-0000-7000-8000-${String(n).padStart(12, "0")}`;
const keys: Record<string, string> = {};

async function makeUser(id: string, plan?: "pro"): Promise<string> {
	await db.insert(user).values({ id, name: id, email: `${id}@example.test`, emailVerified: true });
	if (plan) {
		await db.insert(subscriptions).values({
			userId: id,
			plan,
			provider: "stripe",
			status: "active",
			currentPeriodEnd: new Date(Date.now() + 30 * 86_400_000),
		});
	}
	const k = generateApiKey();
	await db.insert(apiKeys).values({ userId: id, name: "test", prefix: k.prefix, hash: k.hash });
	return k.key;
}

beforeAll(async () => {
	({ db } = await createTestDatabase());
	mock.module("@/lib/db/client", () => ({ db }));
	routes = {
		list: (await import("@/app/api/v1/[type]/route")).GET,
		detail: (await import("@/app/api/v1/[type]/[id]/route")).GET,
		genres: (await import("@/app/api/v1/[type]/genres/route")).GET,
		search: (await import("@/app/api/v1/search/route")).GET,
		lookup: (await import("@/app/api/v1/lookup/route")).GET,
		usage: (await import("@/app/api/v1/me/usage/route")).GET,
	};
	const s = getStore();
	if (!s) throw new Error("expected the in-memory quota store in tests");
	store = s;
	const realThrottle = store.throttle.bind(store);
	store.throttle = (key, perSecond) => (throttle.off ? Promise.resolve(0) : realThrottle(key, perSecond));

	const games = [
		{
			slug: "the-legend-of-zelda",
			title: "The Legend of Zelda",
			date: "1986-02-21",
			score: 90,
			genres: ["Adventure"],
		},
		{ slug: "space-rally", title: "Space Rally", date: "2020-05-01", score: 60, genres: ["Racing"] },
		{ slug: "neon-drift", title: "Neon Drift", date: "2023-03-03", score: 75, genres: ["Racing"] },
	];
	for (const [i, g] of games.entries()) {
		await db.insert(items).values({
			id: uuid(i + 1),
			type: "game",
			slug: g.slug,
			wikidataQid: `Q${i + 1}`,
			releaseDate: g.date,
			popularity: 10 - i,
			ratingCount: 10,
			score: g.score,
			scoreC1: g.score,
			scoreC2: g.score - 10,
			scoreC3: g.score,
			scoreC4: g.score,
			scoreC5: g.score,
			scoreC6: g.score,
			meta: {
				genres: g.genres,
				platforms: ["PC"],
				trailers: [{ youtubeId: "AAAAAAAAAAA" }],
				creators: ["Studio"],
			},
			externalIds: i === 0 ? { steam: "123456", imdb: "tt0000001" } : {},
		});
		await db.insert(itemTranslations).values({
			itemId: uuid(i + 1),
			locale: "en",
			title: g.title,
			synopsis: `${g.title} synopsis`,
			synopsisLicense: "CC BY-SA 4.0",
			synopsisSourceUrl: "https://en.wikipedia.org/wiki/X",
		});
	}
	await db.insert(itemTranslations).values({ itemId: uuid(1), locale: "pt-BR", title: "A Lenda de Zelda" });
	await db.insert(items).values({ id: uuid(99), type: "movie", slug: "a-movie", wikidataQid: "Q99", ratingCount: 2 });
	await db.insert(itemTranslations).values({ itemId: uuid(99), locale: "en", title: "A Movie" });
	await db.insert(externalScores).values({ itemId: uuid(1), source: "metacritic", value: "88" });

	keys.free = await makeUser("u_free");
	keys.pro = await makeUser("u_pro", "pro");
	keys.quota = await makeUser("u_quota");
	keys.throttled = await makeUser("u_throttled");
	keys.revoked = await makeUser("u_revoked");
	await db.update(apiKeys).set({ revokedAt: new Date() }).where(eq(apiKeys.userId, "u_revoked"));
}, 60_000);

beforeEach(() => {
	throttle.off = true;
});

function call(
	route: Route,
	path: string,
	params: Record<string, string> = {},
	key: string | null = keys.free ?? null,
): Promise<Response> {
	const headers: Record<string, string> = key ? { authorization: `Bearer ${key}` } : {};
	return route(new Request(`http://localhost:3000/api/v1${path}`, { headers }), {
		params: Promise.resolve(params),
	});
}

async function expectProblem(res: Response, status: number, code: string): Promise<z.infer<typeof problemSchema>> {
	expect(res.status).toBe(status);
	expect(res.headers.get("content-type")).toContain("application/problem+json");
	const body = problemSchema.parse(await res.json());
	expect(body.code).toBe(code as never);
	expect(body.type).toEndWith(`/docs/errors#${code}`);
	expect(body.request_id).toBe(res.headers.get("x-request-id") as string);
	return body;
}

describe("auth", () => {
	test("401 without a key", async () => {
		const res = await call(routes.list, "/games", { type: "games" }, null);
		await expectProblem(res, 401, "unauthorized");
		expect(res.headers.get("www-authenticate")).toContain("Bearer");
	});

	test("401 with an unknown or revoked key", async () => {
		await expectProblem(
			await call(routes.list, "/games", { type: "games" }, `hx_live_${"x".repeat(32)}`),
			401,
			"unauthorized",
		);
		await expectProblem(await call(routes.list, "/games", { type: "games" }, keys.revoked), 401, "unauthorized");
	});
});

describe("validation", () => {
	test("400 lists every invalid parameter", async () => {
		const res = await call(routes.list, "/games?sort=name&min_score=101&min_acting=5", { type: "games" });
		const body = await expectProblem(res, 400, "bad_request");
		expect(body.errors?.map((e) => e.path).sort()).toEqual(["", "min_score", "sort"]);
		expect(JSON.stringify(body.errors)).toContain("min_acting");
	});

	test("empty query values are rejected instead of coerced to 0", async () => {
		for (const query of ["min_score=", "limit=", "year=&genre=Action"]) {
			const res = await call(routes.list, `/games?${query}`, { type: "games" });
			const body = await expectProblem(res, 400, "bad_request");
			expect(JSON.stringify(body)).toContain("Must not be empty.");
		}
	});

	test("limit=11 is rejected on Free and accepted on Pro", async () => {
		const res = await call(routes.list, "/games?limit=11", { type: "games" });
		const body = await expectProblem(res, 400, "bad_request");
		expect(body.errors?.[0]?.message).toContain("at most 10");
		expect((await call(routes.list, "/games?limit=11", { type: "games" }, keys.pro)).status).toBe(200);
	});

	test("unknown type or malformed id is 404", async () => {
		await expectProblem(await call(routes.list, "/cars", { type: "cars" }), 404, "not_found");
		await expectProblem(await call(routes.detail, "/games/1", { type: "games", id: "1" }), 404, "not_found");
	});

	test("a tampered cursor is 400", async () => {
		await expectProblem(await call(routes.list, "/games?cursor=abc", { type: "games" }), 400, "bad_request");
	});

	test("search needs at least 2 characters", async () => {
		await expectProblem(await call(routes.search, "/search?q=z"), 400, "bad_request");
	});
});

describe("quota and rate limit", () => {
	test("Free gets 100 requests, then 402", async () => {
		for (let i = 0; i < 100; i++) {
			expect((await call(routes.usage, "/me/usage", {}, keys.quota)).status).toBe(200);
		}
		const res = await call(routes.usage, "/me/usage", {}, keys.quota);
		expect(await expectProblem(res, 402, "payment_required")).toMatchObject({ plan: "free", limit: 100 });
	});

	test("second request in the same second is 429 with Retry-After", async () => {
		throttle.off = false;
		const first = await call(routes.genres, "/games/genres", { type: "games" }, keys.throttled);
		const second = await call(routes.genres, "/games/genres", { type: "games" }, keys.throttled);
		// The pair can straddle a second boundary; one of the next two must then be throttled.
		const third = await call(routes.genres, "/games/genres", { type: "games" }, keys.throttled);
		expect(first.status).toBe(200);
		const throttled = [second, third].find((r) => r.status === 429);
		expect(throttled).toBeDefined();
		await expectProblem(throttled as Response, 429, "rate_limited");
		expect(Number(throttled?.headers.get("retry-after"))).toBeGreaterThan(0);
	});
});

test("a failing rate-limit store fails closed with 503", async () => {
	const original = store.incr;
	store.incr = () => Promise.reject(new Error("redis down"));
	try {
		const res = await call(routes.genres, "/games/genres", { type: "games" });
		await expectProblem(res, 503, "service_unavailable");
		expect(res.headers.get("retry-after")).toBe("30");
	} finally {
		store.incr = original;
	}
});

describe("endpoints", () => {
	test("list: shape, headers, filters, sort and cursor", async () => {
		const res = await call(routes.list, "/games?sort=-score&limit=2", { type: "games" });
		expect(res.status).toBe(200);
		for (const h of [
			"x-request-id",
			"ratelimit-limit",
			"ratelimit-remaining",
			"ratelimit-reset",
			"x-quota-remaining",
		]) {
			expect(res.headers.get(h)).not.toBeNull();
		}
		expect(res.headers.get("cache-control")).toBe("private, no-store");
		const page = itemListSchema.parse(await res.json());
		expect(page.data.map((i) => i.slug)).toEqual(["the-legend-of-zelda", "neon-drift"]);
		expect(page.data[0]?.categories).toMatchObject({ gameplay: 90, fun: 80 });
		expect(page.data[0]?.url).toEndWith("/games/the-legend-of-zelda");

		const next = itemListSchema.parse(
			await call(routes.list, `/games?sort=-score&limit=2&cursor=${page.next_cursor}`, { type: "games" }).then(
				(r) => r.json(),
			),
		);
		expect(next.data.map((i) => i.slug)).toEqual(["space-rally"]);
		expect(next.next_cursor).toBeNull();

		const filtered = itemListSchema.parse(
			await call(routes.list, "/games?genre=Racing&min_fun=60", { type: "games" }).then((r) => r.json()),
		);
		expect(filtered.data.map((i) => i.slug)).toEqual(["neon-drift"]);
	});

	test("list in pt-BR translates titles and links", async () => {
		const page = itemListSchema.parse(
			await call(routes.list, "/games?locale=pt-BR&sort=-score", { type: "games" }).then((r) => r.json()),
		);
		expect(page.data[0]?.title).toBe("A Lenda de Zelda");
		expect(page.data[0]?.url).toContain("/pt-br/games/");
	});

	test("detail: full item, wrong type and unknown id are 404", async () => {
		const res = await call(routes.detail, `/games/${uuid(1)}`, { type: "games", id: uuid(1) });
		expect(res.status).toBe(200);
		const item = itemDetailSchema.parse(await res.json());
		expect(item.external_ids).toMatchObject({ wikidata: "Q1", steam: "123456", imdb: "tt0000001" });
		expect(item.trailers).toEqual([{ youtube_id: "AAAAAAAAAAA", title: null }]);
		expect(item.synopsis).toEqual({
			text: "The Legend of Zelda synopsis",
			locale: "en",
			license: "CC BY-SA 4.0",
			source_url: "https://en.wikipedia.org/wiki/X",
		});
		await expectProblem(
			await call(routes.detail, `/movies/${uuid(1)}`, { type: "movies", id: uuid(1) }),
			404,
			"not_found",
		);
		await expectProblem(
			await call(routes.detail, `/games/${uuid(50)}`, { type: "games", id: uuid(50) }),
			404,
			"not_found",
		);
	});

	test("detail hides scores under 5 ratings", async () => {
		const item = itemDetailSchema.parse(
			await call(routes.detail, `/movies/${uuid(99)}`, { type: "movies", id: uuid(99) }).then((r) => r.json()),
		);
		expect(item.score).toBeNull();
		expect(Object.values(item.categories).every((v) => v === null)).toBe(true);
		expect(item.rating_count).toBe(2);
	});

	test("detail: time_to_beat on games (median from 3 reports), null on other types", async () => {
		const reporters = ["ttb1", "ttb2", "ttb3"];
		await db
			.insert(user)
			.values(reporters.map((id) => ({ id, name: id, email: `${id}@example.test`, emailVerified: true })));
		await db.insert(timeToBeat).values(
			reporters.map((userId, i) => ({
				userId,
				itemId: uuid(1),
				mainMinutes: (10 + i * 5) * 60,
				extrasMinutes: i === 0 ? 1800 : null,
			})),
		);
		const game = itemDetailSchema.parse(
			await call(routes.detail, `/games/${uuid(1)}`, { type: "games", id: uuid(1) }).then((r) => r.json()),
		);
		expect(game.time_to_beat).toEqual({
			main: { median_hours: 15, reports: 3 },
			extras: { median_hours: null, reports: 1 },
			completionist: { median_hours: null, reports: 0 },
		});
		const movie = itemDetailSchema.parse(
			await call(routes.detail, `/movies/${uuid(99)}`, { type: "movies", id: uuid(99) }).then((r) => r.json()),
		);
		expect(movie.time_to_beat).toBeNull();
	});

	test("genres with counts", async () => {
		const res = await call(routes.genres, "/games/genres", { type: "games" });
		expect(await res.json()).toEqual({
			data: [
				{ slug: "Racing", count: 2 },
				{ slug: "Adventure", count: 1 },
			],
		});
	});

	test("search across types", async () => {
		const body = itemListSchema
			.pick({ data: true })
			.parse(await call(routes.search, "/search?q=zelda").then((r) => r.json()));
		expect(body.data[0]?.slug).toBe("the-legend-of-zelda");
	});

	test("lookup by external id and by wikidata", async () => {
		const bySteam = itemDetailSchema.parse(
			await call(routes.lookup, "/lookup?source=steam&id=123456").then((r) => r.json()),
		);
		expect(bySteam.id).toBe(uuid(1));
		const byQid = itemDetailSchema.parse(
			await call(routes.lookup, "/lookup?source=wikidata&id=Q2").then((r) => r.json()),
		);
		expect(byQid.slug).toBe("space-rally");
		await expectProblem(await call(routes.lookup, "/lookup?source=imdb&id=tt9999999"), 404, "not_found");
		await expectProblem(await call(routes.lookup, "/lookup?source=metacritic&id=x"), 400, "bad_request");
	});

	test("me/usage reports plan, quota and rate limit", async () => {
		const usage = usageSchema.parse(await call(routes.usage, "/me/usage", {}, keys.pro).then((r) => r.json()));
		expect(usage).toMatchObject({
			plan: "pro",
			quota: { limit: 50_000, period: "month" },
			credits: { balance: 0 },
			rate_limit: { requests_per_second: 10 },
		});
		expect(usage.quota.used).toBeGreaterThan(0);
		expect(usage.quota.resets_at).not.toBeNull();
	});

	test("third-party scores never leave the API", async () => {
		const bodies = await Promise.all([
			call(routes.detail, `/games/${uuid(1)}`, { type: "games", id: uuid(1) }).then((r) => r.text()),
			call(routes.list, "/games", { type: "games" }).then((r) => r.text()),
			call(routes.lookup, "/lookup?source=steam&id=123456").then((r) => r.text()),
		]);
		for (const body of bodies) {
			expect(body).not.toContain("external_scores");
			expect(body).not.toContain("metacritic");
		}
	});
});
