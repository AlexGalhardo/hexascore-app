import { afterAll, beforeAll, describe, expect, test } from "bun:test";
import { eq } from "drizzle-orm";
import { purgeScheduledAccounts } from "./account-purge";
import { getItemById } from "./catalog/repository";
import type { Database } from "./db/client";
import { user } from "./db/schema/auth";
import { items, itemTranslations } from "./db/schema/catalog";
import { createTestDatabase } from "./db/testing";
import { asMember } from "./member";
import { getUserRating, ratingInputSchema, removeRating, writeRating } from "./ratings";
import type { SessionUser } from "./session";

let db: Database;
let close: () => Promise<void>;
const ITEM = "01900000-0000-7000-8000-000000000001";
const OTHER = "01900000-0000-7000-8000-000000000002";
const MISSING = "01900000-0000-7000-8000-0000000000ff";

const stats = async (id = ITEM) => {
	const [row] = await db.select().from(items).where(eq(items.id, id));
	if (!row) throw new Error("item missing");
	return row;
};

beforeAll(async () => {
	({ db, close } = await createTestDatabase());
	for (const [i, id] of [ITEM, OTHER].entries()) {
		await db.insert(items).values({ id, type: "game", slug: `game-${i}`, wikidataQid: `Q${i}` });
		await db.insert(itemTranslations).values({ itemId: id, locale: "en", title: `Game ${i}` });
	}
	await db.insert(user).values(
		Array.from({ length: 12 }, (_, i) => ({
			id: `u${i}`,
			name: `User ${i}`,
			email: `u${i}@example.com`,
			emailVerified: true,
			username: `user${i}`,
		})),
	);
});

afterAll(async () => {
	await close();
});

describe("ratingInputSchema", () => {
	test("needs exactly 6 integers 0..100", () => {
		const ok = { itemId: ITEM, scores: [0, 100, 50, 50, 50, 50] };
		expect(ratingInputSchema.safeParse(ok).success).toBe(true);
		for (const scores of [
			[1, 2, 3, 4, 5],
			[1, 2, 3, 4, 5, 6, 7],
			[101, 0, 0, 0, 0, 0],
			[-1, 0, 0, 0, 0, 0],
			[50.5, 0, 0, 0, 0, 0],
			["50", 0, 0, 0, 0, 0],
		]) {
			expect(ratingInputSchema.safeParse({ itemId: ITEM, scores }).success).toBe(false);
		}
		expect(ratingInputSchema.safeParse({ itemId: "nope", scores: ok.scores }).success).toBe(false);
	});
});

describe("asMember", () => {
	const base: SessionUser = {
		id: "x",
		name: "X",
		email: "x@example.com",
		username: "xuser",
		emailVerified: true,
		role: null,
		banned: false,
		image: null,
		twoFactorEnabled: false,
		profilePrivate: false,
		deletionScheduledAt: null,
	};
	test("rejects anonymous, banned, unverified and nickname-less users", () => {
		expect(asMember(null)).toBe("unauthorized");
		expect(asMember({ ...base, banned: true })).toBe("forbidden");
		expect(asMember({ ...base, emailVerified: false })).toBe("emailNotVerified");
		expect(asMember({ ...base, username: null })).toBe("nicknameRequired");
		expect(asMember(base)).toMatchObject({ id: "x", username: "xuser" });
	});
});

describe("rating write path", () => {
	test("insert, edit and delete keep the stats in sync", async () => {
		expect(await writeRating(db, "u0", { itemId: ITEM, scores: [80, 80, 80, 80, 80, 81] })).toBe(true);
		let s = await stats();
		expect(s.ratingCount).toBe(1);
		expect(s.score).toBe(80);
		expect(s.scoreC6).toBe(81);
		expect(await getUserRating(db, "u0", ITEM)).toEqual([80, 80, 80, 80, 80, 81]);

		await writeRating(db, "u0", { itemId: ITEM, scores: [10, 20, 30, 40, 50, 60] });
		s = await stats();
		expect(s.ratingCount).toBe(1);
		expect(s.score).toBe(35); // round(210 / 6) = 35
		expect([s.scoreC1, s.scoreC2, s.scoreC6]).toEqual([10, 20, 60]);

		await writeRating(db, "u1", { itemId: ITEM, scores: [11, 21, 31, 41, 51, 61] });
		s = await stats();
		expect(s.ratingCount).toBe(2);
		expect(s.score).toBe(36); // users' overall 35 and 36 → 35.5 → 36
		expect(s.scoreC1).toBe(11); // 10.5 → 11 (half away from zero)

		expect(await removeRating(db, "u1", ITEM)).toBe(true);
		expect(await removeRating(db, "u0", ITEM)).toBe(true);
		s = await stats();
		expect(s.ratingCount).toBe(0);
		expect(s.score).toBeNull();
		expect(s.scoreC1).toBeNull();
		expect(await getUserRating(db, "u0", ITEM)).toBeNull();
	});

	test("unknown item writes nothing", async () => {
		expect(await writeRating(db, "u0", { itemId: MISSING, scores: [1, 1, 1, 1, 1, 1] })).toBe(false);
		expect(await removeRating(db, "u0", MISSING)).toBe(false);
	});

	test("concurrent writes from many users do not lose updates", async () => {
		await Promise.all(
			Array.from({ length: 10 }, (_, i) =>
				writeRating(db, `u${i}`, { itemId: OTHER, scores: [i * 10, i * 10, i * 10, i * 10, i * 10, i * 10] }),
			),
		);
		const s = await stats(OTHER);
		expect(s.ratingCount).toBe(10);
		expect(s.score).toBe(45); // mean of 0, 10, …, 90
		expect(s.scoreC3).toBe(45);
	});

	test("metascore becomes visible at the 5th rating", async () => {
		for (let i = 0; i < 4; i++) await writeRating(db, `u${i}`, { itemId: ITEM, scores: [70, 70, 70, 70, 70, 70] });
		let detail = await getItemById(db, ITEM, "en");
		expect(detail?.ratingCount).toBe(4);
		expect(detail?.score).toBeNull();
		expect((await stats()).score).toBe(70); // stored raw, hidden by the read side

		await writeRating(db, "u4", { itemId: ITEM, scores: [100, 100, 100, 100, 100, 100] });
		detail = await getItemById(db, ITEM, "en");
		expect(detail?.ratingCount).toBe(5);
		expect(detail?.score).toBe(76); // (4 × 70 + 100) / 5
		expect(detail?.categories).toEqual([76, 76, 76, 76, 76, 76]);
	});

	test("ratings of a deleted (anonymized) account still count", async () => {
		const before = await stats();
		await db
			.update(user)
			.set({ deletionScheduledAt: new Date("2000-01-01") })
			.where(eq(user.id, "u4"));
		expect(await purgeScheduledAccounts(db)).toEqual(["u4"]);
		await writeRating(db, "u5", { itemId: ITEM, scores: [70, 70, 70, 70, 70, 70] });
		const after = await stats();
		expect(after.ratingCount).toBe(before.ratingCount + 1);
		expect(after.score).toBe(75); // (5 × 70 + 100) / 6 = 75
	});
});
