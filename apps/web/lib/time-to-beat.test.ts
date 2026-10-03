import { afterAll, beforeAll, describe, expect, test } from "bun:test";
import { mergeItems } from "./admin";
import { getItemById } from "./catalog/repository";
import type { Database } from "./db/client";
import { user } from "./db/schema/auth";
import { items, itemTranslations } from "./db/schema/catalog";
import { createTestDatabase } from "./db/testing";
import { setItemStatus } from "./social/repository";
import {
	getTimeToBeat,
	getUserTimeToBeat,
	removeTimeToBeat,
	timeToBeatInputSchema,
	writeTimeToBeat,
} from "./time-to-beat";

let db: Database;
let close: () => Promise<void>;
const GAME = "01900000-0000-7000-8000-000000000001";
const MOVIE = "01900000-0000-7000-8000-000000000002";
const DUPLICATE = "01900000-0000-7000-8000-000000000003";
const MISSING = "01900000-0000-7000-8000-0000000000ff";

const report = (userId: string, main: number | null, extras: number | null = null, itemId = GAME) =>
	writeTimeToBeat(db, userId, { itemId, main, extras, completionist: null });

beforeAll(async () => {
	({ db, close } = await createTestDatabase());
	const rows = [
		{ id: GAME, type: "game" as const, slug: "game" },
		{ id: MOVIE, type: "movie" as const, slug: "movie" },
		{ id: DUPLICATE, type: "game" as const, slug: "game-dup" },
	];
	for (const [i, row] of rows.entries()) {
		await db.insert(items).values({ ...row, wikidataQid: `Q${i}` });
		await db.insert(itemTranslations).values({ itemId: row.id, locale: "en", title: row.slug });
	}
	await db.insert(user).values(
		Array.from({ length: 6 }, (_, i) => ({
			id: `u${i}`,
			name: `U${i}`,
			email: `u${i}@example.com`,
			emailVerified: true,
		})),
	);
	for (let i = 0; i < 6; i++) await setItemStatus(db, `u${i}`, GAME, "completed");
});

afterAll(async () => {
	await close();
});

describe("timeToBeatInputSchema", () => {
	test("hours in 0.5 steps within 0.5..10000, at least one style", () => {
		const base = { itemId: GAME, main: null, extras: null, completionist: null };
		expect(timeToBeatInputSchema.safeParse({ ...base, main: 0.5 }).success).toBe(true);
		expect(timeToBeatInputSchema.safeParse({ ...base, completionist: 10_000 }).success).toBe(true);
		for (const main of [0, 0.25, 12.3, 10_000.5, -1, "10"]) {
			expect(timeToBeatInputSchema.safeParse({ ...base, main }).success).toBe(false);
		}
		expect(timeToBeatInputSchema.safeParse(base).success).toBe(false);
		expect(timeToBeatInputSchema.safeParse({ ...base, itemId: "nope", main: 1 }).success).toBe(false);
	});
});

describe("write path", () => {
	test("only completed games accept reports", async () => {
		expect(await report("u0", 10, null, MISSING)).toBe("notFound");
		await setItemStatus(db, "u0", MOVIE, "completed");
		expect(await report("u0", 10, null, MOVIE)).toBe("forbidden");
		await setItemStatus(db, "u0", DUPLICATE, "in_progress");
		expect(await report("u0", 10, null, DUPLICATE)).toBe("forbidden");
		expect(await report("u0", 10)).toBe("ok");
		expect(await getUserTimeToBeat(db, "u0", GAME)).toEqual({ main: 10, extras: null, completionist: null });
	});

	test("edit, delete, and leaving completed removes the report", async () => {
		expect(await report("u0", 12.5, 20)).toBe("ok");
		expect(await getUserTimeToBeat(db, "u0", GAME)).toEqual({ main: 12.5, extras: 20, completionist: null });
		await removeTimeToBeat(db, "u0", GAME);
		expect(await getUserTimeToBeat(db, "u0", GAME)).toBeNull();

		expect(await report("u0", 10)).toBe("ok");
		expect(await setItemStatus(db, "u0", GAME, "completed")).toBe("ok");
		expect(await setItemStatus(db, "u0", GAME, "dropped")).toBe("cleared");
		expect(await getUserTimeToBeat(db, "u0", GAME)).toBeNull();
		expect(await setItemStatus(db, "u0", GAME, "completed")).toBe("ok");
	});
});

describe("aggregation", () => {
	test("median per style, hidden below 3 reports", async () => {
		expect(await getTimeToBeat(db, GAME)).toEqual({
			main: { medianHours: null, reports: 0 },
			extras: { medianHours: null, reports: 0 },
			completionist: { medianHours: null, reports: 0 },
		});
		await report("u1", 10, 20);
		await report("u2", 20, 30);
		let stats = await getTimeToBeat(db, GAME);
		expect(stats.main).toEqual({ medianHours: null, reports: 2 });

		await report("u3", 31);
		await report("u4", 40.5);
		stats = await getTimeToBeat(db, GAME);
		// percentile_cont interpolates: median of 10, 20, 31, 40.5 = 25.5; extras keeps its own 2 reports.
		expect(stats.main).toEqual({ medianHours: 25.5, reports: 4 });
		expect(stats.extras).toEqual({ medianHours: null, reports: 2 });
		expect(stats.completionist.reports).toBe(0);

		const detail = await getItemById(db, GAME, "en");
		expect(detail?.timeToBeat?.main.medianHours).toBe(25.5);
		expect((await getItemById(db, MOVIE, "en"))?.timeToBeat).toBeNull();
	});

	test("merge moves reports only where the merged status is completed", async () => {
		await setItemStatus(db, "u5", DUPLICATE, "completed");
		await report("u5", 50, null, DUPLICATE);
		await setItemStatus(db, "u0", DUPLICATE, "completed");
		await report("u0", 60, null, DUPLICATE);
		// u0 is "completed" on both, u5 only on the duplicate: both statuses and reports land on GAME.
		expect(await mergeItems(db, DUPLICATE, GAME)).toBe("ok");
		const stats = await getTimeToBeat(db, GAME);
		expect(stats.main.reports).toBe(6);
		expect(stats.main.medianHours).toBe(35.8);
	});
});
