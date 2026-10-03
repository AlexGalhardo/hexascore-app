import { afterAll, beforeAll, describe, expect, spyOn, test } from "bun:test";
import type { Database } from "@/lib/db/client";
import { user } from "@/lib/db/schema/auth";
import { affiliateLinks, externalScores, items, itemTranslations } from "@/lib/db/schema/catalog";
import { ratings } from "@/lib/db/schema/social";
import { createTestDatabase } from "@/lib/db/testing";
import { InvalidCursorError } from "./cursor";
import {
	getAffiliateLinks,
	getExternalScores,
	getItemById,
	getItemBySlug,
	getItemRatingsPage,
	listGenres,
	listItems,
	listPlatforms,
	searchItems,
} from "./repository";

let db: Database;
let close: () => Promise<void>;
const ids: Record<string, string> = {};

const uuid = (n: number) => `01900000-0000-7000-8000-${String(n).padStart(12, "0")}`;

beforeAll(async () => {
	({ db, close } = await createTestDatabase());

	const games = [
		{
			slug: "the-legend-of-zelda",
			title: "The Legend of Zelda",
			date: "1986-02-21",
			genres: ["Adventure"],
			score: 90,
		},
		{ slug: "space-rally", title: "Space Rally", date: "2020-05-01", genres: ["Racing"], score: 60 },
		{ slug: "dungeon-deep", title: "Dungeon Deep", date: "2020-11-11", genres: ["RPG", "Adventure"], score: 40 },
		{ slug: "quiet-farm", title: "Quiet Farm", date: null, genres: ["Simulation"], score: null },
		{ slug: "neon-drift", title: "Neon Drift", date: "2023-03-03", genres: ["Racing"], score: 75 },
	];
	for (const [i, g] of games.entries()) {
		const id = uuid(i + 1);
		ids[g.slug] = id;
		await db.insert(items).values({
			id,
			type: "game",
			slug: g.slug,
			wikidataQid: `Q${i + 1}`,
			releaseDate: g.date,
			popularity: 10 - i,
			ratingCount: g.score === null ? 2 : 10,
			score: g.score ?? 55,
			scoreC1: g.score,
			scoreC2: g.score,
			scoreC3: g.score,
			scoreC4: g.score,
			scoreC5: g.score,
			scoreC6: g.score === null ? null : g.score + 5,
			// one deliberately broken field per kind to exercise the lenient jsonb parsing
			meta: {
				genres: g.genres,
				platforms: i === 0 ? ["NES", 42] : ["PC"],
				trailers: [{ youtubeId: "AAAAAAAAAAA" }, { youtubeId: "bad" }],
				runtimeMinutes: "long",
			} as never,
			externalIds: { imdb: "tt0000001", steam: "javascript:alert(1)" } as never,
		});
		await db.insert(itemTranslations).values({
			itemId: id,
			locale: "en",
			title: g.title,
			synopsis: `${g.title} synopsis`,
			synopsisLicense: "CC BY-SA 4.0",
			synopsisSourceUrl: "https://en.wikipedia.org/wiki/X",
		});
	}
	await db.insert(itemTranslations).values({
		itemId: ids["the-legend-of-zelda"] as string,
		locale: "pt-BR",
		title: "A Lenda de Zelda",
		synopsis: "Sinopse",
	});
	await db.insert(items).values({ id: uuid(99), type: "movie", slug: "zelda-the-movie", wikidataQid: "Q99" });
	await db.insert(itemTranslations).values({ itemId: uuid(99), locale: "en", title: "Zelda the Movie" });

	const zelda = ids["the-legend-of-zelda"] as string;
	await db.insert(externalScores).values({ itemId: zelda, source: "metacritic", value: "88" });
	await db.insert(affiliateLinks).values([
		{ itemId: zelda, store: "amazon", url: "https://example.test/a" },
		{ itemId: zelda, store: "gog", url: "https://example.test/g", active: false },
	]);

	for (let u = 1; u <= 25; u++) {
		await db.insert(user).values({
			id: `u${u}`,
			name: `User ${u}`,
			email: `u${u}@example.test`,
			username: `user${u}`,
			profilePrivate: u % 5 === 0,
		});
		await db.insert(ratings).values({
			userId: `u${u}`,
			itemId: zelda,
			c1: u,
			c2: u,
			c3: u,
			c4: u,
			c5: u,
			c6: u,
			createdAt: new Date(Date.UTC(2026, 0, 1, 0, 0, u % 3)),
		});
	}
});

afterAll(async () => {
	await close();
});

describe("listItems", () => {
	test("sorts by visible score descending, items without a score last", async () => {
		const page = await listItems(db, { type: "game", locale: "en", sort: "-score" });
		expect(page.data.map((i) => i.slug)).toEqual([
			"the-legend-of-zelda",
			"neon-drift",
			"space-rally",
			"dungeon-deep",
			"quiet-farm",
		]);
		expect(page.data.at(-1)?.score).toBeNull();
		expect(page.nextCursor).toBeNull();
		expect(page.prevCursor).toBeNull();
	});

	test("ascending score still keeps missing scores last", async () => {
		const page = await listItems(db, { type: "game", locale: "en", sort: "score" });
		expect(page.data[0]?.slug).toBe("dungeon-deep");
		expect(page.data.at(-1)?.slug).toBe("quiet-farm");
	});

	test("keyset pagination walks forward and back without gaps", async () => {
		const seen: string[] = [];
		let page = await listItems(db, { type: "game", locale: "en", sort: "-release_date", limit: 2 });
		seen.push(...page.data.map((i) => i.slug));
		const firstPage = page.data.map((i) => i.slug);
		page = await listItems(db, {
			type: "game",
			locale: "en",
			sort: "-release_date",
			limit: 2,
			cursor: page.nextCursor ?? "",
		});
		const secondPage = page.data.map((i) => i.slug);
		seen.push(...secondPage);
		const back = await listItems(db, {
			type: "game",
			locale: "en",
			sort: "-release_date",
			limit: 2,
			cursor: page.prevCursor ?? "",
		});
		expect(back.data.map((i) => i.slug)).toEqual(firstPage);
		expect(back.prevCursor).toBeNull();
		page = await listItems(db, {
			type: "game",
			locale: "en",
			sort: "-release_date",
			limit: 2,
			cursor: page.nextCursor ?? "",
		});
		seen.push(...page.data.map((i) => i.slug));
		expect(page.nextCursor).toBeNull();
		expect(seen).toEqual(["neon-drift", "dungeon-deep", "space-rally", "the-legend-of-zelda", "quiet-farm"]);
	});

	test("filters by genre, platform, year, score range and category", async () => {
		const slugs = async (extra: object) =>
			(await listItems(db, { type: "game", locale: "en", sort: "popularity", ...extra })).data.map((i) => i.slug);
		expect(await slugs({ genre: "Adventure" })).toEqual(["dungeon-deep", "the-legend-of-zelda"]);
		expect(await slugs({ platform: "NES" })).toEqual(["the-legend-of-zelda"]);
		expect(await slugs({ year: 2020 })).toEqual(["dungeon-deep", "space-rally"]);
		expect(await slugs({ minScore: 60, maxScore: 80 })).toEqual(["neon-drift", "space-rally"]);
		expect(await slugs({ minCategory: { c6: 80 } })).toEqual(["neon-drift", "the-legend-of-zelda"]);
		expect(await slugs({ minScore: 0 })).not.toContain("quiet-farm");
	});

	test("rejects a tampered cursor or one from another sort", async () => {
		const page = await listItems(db, { type: "game", locale: "en", limit: 1 });
		expect(listItems(db, { type: "game", locale: "en", cursor: "not-base64!" })).rejects.toBeInstanceOf(
			InvalidCursorError,
		);
		expect(
			listItems(db, { type: "game", locale: "en", sort: "-score", cursor: page.nextCursor ?? "" }),
		).rejects.toBeInstanceOf(InvalidCursorError);
	});

	test("uses the locale title with English fallback", async () => {
		const page = await listItems(db, { type: "game", locale: "pt-BR", sort: "-score", limit: 2 });
		expect(page.data.map((i) => i.title)).toEqual(["A Lenda de Zelda", "Neon Drift"]);
	});
});

describe("item detail", () => {
	test("returns parsed meta and drops invalid jsonb parts", async () => {
		const warn = spyOn(console, "warn").mockImplementation(() => {});
		const item = await getItemBySlug(db, "game", "the-legend-of-zelda", "en");
		expect(item?.meta.platforms).toEqual(["NES"]);
		expect(item?.meta.trailers).toEqual([{ youtubeId: "AAAAAAAAAAA" }]);
		expect(item?.meta.runtimeMinutes).toBeUndefined();
		expect(item?.externalIds).toEqual({ imdb: "tt0000001" });
		expect(item?.score).toBe(90);
		expect(item?.translationFallback).toBe(false);
		expect(warn).toHaveBeenCalled();
		warn.mockRestore();
	});

	test("flags the English fallback", async () => {
		const es = await getItemBySlug(db, "game", "space-rally", "es");
		expect(es?.title).toBe("Space Rally");
		expect(es?.synopsis).toBe("Space Rally synopsis");
		expect(es?.translationFallback).toBe(true);
		const pt = await getItemBySlug(db, "game", "the-legend-of-zelda", "pt-BR");
		expect(pt?.title).toBe("A Lenda de Zelda");
		expect(pt?.synopsisLicense).toBeNull();
		expect(pt?.translationFallback).toBe(false);
	});

	test("hides the score below 5 ratings and handles unknown ids/slugs", async () => {
		const farm = await getItemBySlug(db, "game", "quiet-farm", "en");
		expect(farm?.score).toBeNull();
		expect(farm?.categories).toEqual([null, null, null, null, null, null]);
		expect(await getItemBySlug(db, "movie", "the-legend-of-zelda", "en")).toBeNull();
		expect(await getItemById(db, "not-a-uuid", "en")).toBeNull();
		expect((await getItemById(db, ids["space-rally"] as string, "en"))?.slug).toBe("space-rally");
	});
});

describe("searchItems", () => {
	test("tolerates typos and groups across types", async () => {
		const results = await searchItems(db, "zleda", { locale: "en" });
		expect(results.map((r) => r.slug)).toContain("the-legend-of-zelda");
		expect(results.map((r) => r.slug)).toContain("zelda-the-movie");
	});

	test("filters by type and matches localized titles", async () => {
		const results = await searchItems(db, "lenda", { locale: "pt-BR", type: "game" });
		expect(results.map((r) => r.title)).toEqual(["A Lenda de Zelda"]);
		expect(await searchItems(db, "zelda", { locale: "en", type: "book" })).toEqual([]);
		expect(await searchItems(db, " z ", { locale: "en" })).toEqual([]);
	});
});

describe("side data", () => {
	test("facets, external scores and active affiliate links", async () => {
		expect(await listGenres(db, "game")).toEqual(["Adventure", "RPG", "Racing", "Simulation"]);
		expect(await listPlatforms(db, "game")).toContain("NES");
		const zelda = ids["the-legend-of-zelda"] as string;
		expect(await getExternalScores(db, zelda)).toEqual([{ source: "metacritic", value: "88", url: null }]);
		expect((await getAffiliateLinks(db, zelda)).map((l) => l.store)).toEqual(["amazon"]);
	});
});

describe("getItemRatingsPage", () => {
	test("pages through public ratings only", async () => {
		const zelda = ids["the-legend-of-zelda"] as string;
		const nicknames: string[] = [];
		let cursor: string | undefined;
		do {
			const page = await getItemRatingsPage(db, zelda, { cursor, limit: 7 });
			nicknames.push(...page.data.map((r) => r.nickname));
			cursor = page.nextCursor ?? undefined;
		} while (cursor);
		expect(nicknames).toHaveLength(20);
		expect(new Set(nicknames).size).toBe(20);
		expect(nicknames).not.toContain("user5");
		expect(getItemRatingsPage(db, zelda, { cursor: "%%%" })).rejects.toBeInstanceOf(InvalidCursorError);
	});
});
