import { afterAll, beforeAll, describe, expect, test } from "bun:test";
import { count, eq, like, max, sql } from "drizzle-orm";
import { z } from "zod";
import { REPORT_REASONS } from "@/lib/social/constants";
import { contactSchema } from "@/lib/validation";
import type { Database } from "./client";
import { user } from "./schema/auth";
import { type ItemType, items } from "./schema/catalog";
import { contactMessages, follows, itemStatus, listItems, lists, ratings, reports, timeToBeat } from "./schema/social";
import { purgeSeedUsers, repairSeedIds, seedExistingItems, seedModeration } from "./seed";
import { createTestDatabase } from "./testing";

// --existing-items against a pre-existing catalog (the sandbox case): 400 games with tied popularity (hash tie-break),
// movies with varied popularity, books without popularity data.
let db: Database;
let close: () => Promise<void>;
const USERS = 30;
const OPTS = { users: USERS, batch: 8 };

const catalog = async () => {
	const rows = await db.select().from(items).orderBy(items.id);
	return rows.map(({ ratingCount, score, scoreC1, scoreC2, scoreC3, scoreC4, scoreC5, scoreC6, ...rest }) => rest);
};
const fullRows = () => db.select().from(items).orderBy(items.id);
const counts = async () => ({
	ratings: (await db.select({ n: count() }).from(ratings))[0]?.n,
	statuses: (await db.select({ n: count() }).from(itemStatus))[0]?.n,
	ttb: (await db.select({ n: count() }).from(timeToBeat))[0]?.n,
	follows: (await db.select({ n: count() }).from(follows))[0]?.n,
	lists: (await db.select({ n: count() }).from(lists))[0]?.n,
	listItems: (await db.select({ n: count() }).from(listItems))[0]?.n,
});
const moderation = async () => ({
	banned: (await db.select({ n: count() }).from(user).where(eq(user.banned, true)))[0]?.n ?? -1,
	hiddenLists: (await db.select({ n: count() }).from(lists).where(eq(lists.hidden, true)))[0]?.n ?? -1,
	reports: (await db.select({ n: count() }).from(reports))[0]?.n ?? -1,
	contactMessages: (await db.select({ n: count() }).from(contactMessages))[0]?.n ?? -1,
});
const countWhere = async (table: typeof items | typeof ratings, where: ReturnType<typeof sql>) =>
	(await db.select({ n: count() }).from(table).where(where))[0]?.n ?? -1;

beforeAll(async () => {
	({ db, close } = await createTestDatabase());
	const rows: { type: ItemType; slug: string; wikidataQid: string; popularity: number }[] = [];
	for (let i = 0; i < 400; i++) rows.push({ type: "game", slug: `g-${i}`, wikidataQid: `QG${i}`, popularity: 100 });
	for (let i = 0; i < 20; i++) rows.push({ type: "movie", slug: `m-${i}`, wikidataQid: `QM${i}`, popularity: i * 5 });
	for (let i = 0; i < 10; i++) rows.push({ type: "book", slug: `b-${i}`, wikidataQid: `QB${i}`, popularity: 0 });
	await db.insert(items).values(rows);
}, 60_000);

afterAll(() => close());

describe("seed --existing-items", () => {
	let before: Awaited<ReturnType<typeof catalog>>;
	let firstRun: Awaited<ReturnType<typeof counts>>;
	let moderationFirstRun: Awaited<ReturnType<typeof moderation>>;

	test("rates existing items without touching anything but their stats", async () => {
		before = await catalog();
		const written = await seedExistingItems(db, OPTS);
		expect(written).toBeGreaterThan(500);
		expect(await catalog()).toEqual(before);
		expect((await db.select({ n: count() }).from(user).where(like(user.id, "seed-user-%")))[0]?.n).toBe(USERS);
		expect(await countWhere(ratings, sql`${ratings.userId} not like 'seed-user-%'`)).toBe(0);
		// The most popular item of each type gets 60% of the users; stats match lib/ratings.ts' formula.
		const [hit] = await db.select({ n: max(items.ratingCount) }).from(items);
		expect(hit?.n).toBe(USERS * 0.6);
		const wrongStats = sql`(${items.ratingCount}, ${items.score}, ${items.scoreC1}, ${items.scoreC6}) is distinct from (
			select row(count(*)::int, round(avg(r.overall))::smallint, round(avg(r.c1))::smallint, round(avg(r.c6))::smallint)
			from ratings r where r.item_id = ${items.id})`;
		expect(await countWhere(items, wrongStats)).toBe(0);
		// Popular movies get more raters than obscure ones.
		const [top] = await db
			.select({ n: max(items.ratingCount) })
			.from(items)
			.where(eq(items.slug, "m-19"));
		const [low] = await db
			.select({ n: max(items.ratingCount) })
			.from(items)
			.where(eq(items.slug, "m-1"));
		expect(top?.n ?? 0).toBeGreaterThan(low?.n ?? 0);
	}, 120_000);

	test("social data follows the app rules", async () => {
		const perType = await db
			.select({ n: count() })
			.from(lists)
			.groupBy(lists.userId, lists.type)
			.orderBy(sql`count(*) desc`)
			.limit(1);
		// LIST_LIMITS.perType; the cap itself is only reached at scale (power raters with 225+ ratings of a type).
		expect(perType[0]?.n).toBeGreaterThan(0);
		expect(perType[0]?.n).toBeLessThanOrEqual(10);
		const mismatched = await db
			.select({ n: count() })
			.from(listItems)
			.innerJoin(lists, eq(lists.id, listItems.listId))
			.innerJoin(items, eq(items.id, listItems.itemId))
			.where(sql`${items.type} <> ${lists.type}`);
		expect(mismatched[0]?.n).toBe(0);

		const badTtb = await db
			.select({ n: count() })
			.from(timeToBeat)
			.innerJoin(items, eq(items.id, timeToBeat.itemId))
			.where(
				sql`${items.type} <> 'game' or not exists (select 1 from item_status s where s.user_id = ${timeToBeat.userId}
					and s.item_id = ${timeToBeat.itemId} and s.status = 'completed')`,
			);
		expect(badTtb[0]?.n).toBe(0);

		firstRun = await counts();
		for (const n of Object.values(firstRun)) expect(n).toBeGreaterThan(0);
		const statuses = await db.selectDistinct({ s: itemStatus.status }).from(itemStatus);
		expect(statuses.map((r) => r.s).sort()).toEqual(["completed", "dropped", "in_progress", "want"]);
	});

	test("moderation data follows the app rules", async () => {
		moderationFirstRun = await moderation();
		expect(moderationFirstRun.reports).toBeGreaterThan(0);
		expect(moderationFirstRun.contactMessages).toBeGreaterThan(0);

		// Nobody reports their own profile or their own list, and every reporter is a seed user.
		const selfReports = await db
			.select({ n: count() })
			.from(reports)
			.where(
				sql`${reports.reporterId} not like 'seed-user-%'
					or (${reports.targetType} = 'user' and ${reports.targetId} = ${reports.reporterId})
					or (${reports.targetType} = 'list' and exists (select 1 from lists l
						where l.id::text = ${reports.targetId} and l.user_id = ${reports.reporterId}))`,
			);
		expect(selfReports[0]?.n).toBe(0);
		const reasons = await db.selectDistinct({ r: reports.reason }).from(reports);
		for (const { r } of reasons) expect(REPORT_REASONS as readonly string[]).toContain(r);

		const messages = await db.select().from(contactMessages);
		for (const m of messages) {
			expect(contactSchema.safeParse(m).success).toBe(true);
			expect(m.userId).toMatch(/^seed-user-\d{6}$/);
			expect(m.email).toBe(`${m.userId}@example.test`);
		}
		const bannedWithoutReason = await db
			.select({ n: count() })
			.from(user)
			.where(sql`${user.banned} and ${user.banReason} is null`);
		expect(bannedWithoutReason[0]?.n).toBe(0);
	});

	test("seed ids are valid UUIDs, and ids from older runs are repaired in place", async () => {
		const ids = async () => [
			...(await db.select({ id: lists.id }).from(lists)),
			...(await db.select({ id: reports.id }).from(reports)),
			...(await db.select({ id: contactMessages.id }).from(contactMessages)),
		];
		for (const { id } of await ids()) expect(z.uuid().safeParse(id).success).toBe(true);

		// What earlier versions wrote: bare md5 ids, which z.uuid() rejects.
		const legacyList = "c4ca4238-a0b9-2382-0dcc-509a6f75849b";
		const legacyRow = "c81e728d-9d4c-2f63-6f06-7f89cc14862c";
		const repairedList = "c4ca4238-a0b9-4382-8dcc-509a6f75849b";
		const [item] = await db.select({ id: items.id }).from(items).where(eq(items.slug, "g-1"));
		if (!item) throw new Error("item missing");
		await db.insert(lists).values({ id: legacyList, userId: "seed-user-000001", type: "game", title: "Legacy" });
		await db.insert(listItems).values({ listId: legacyList, itemId: item.id, position: 1 });
		await db.insert(reports).values({
			id: legacyRow,
			reporterId: "seed-user-000002",
			targetType: "list",
			targetId: legacyList,
			reason: "spam",
		});
		const before = await moderation();

		await repairSeedIds(db);
		for (const { id } of await ids()) expect(z.uuid().safeParse(id).success).toBe(true);
		expect(await moderation()).toEqual(before);
		expect(await db.select().from(listItems).where(eq(listItems.listId, repairedList))).toHaveLength(1);
		expect(await db.select().from(reports).where(eq(reports.targetId, repairedList))).toHaveLength(1);

		await db.delete(lists).where(eq(lists.id, repairedList));
		await db.delete(reports).where(eq(reports.targetId, repairedList));
	});

	test("a second run is idempotent", async () => {
		const rows = await fullRows();
		await seedExistingItems(db, OPTS);
		expect(await counts()).toEqual(firstRun);
		expect(await moderation()).toEqual(moderationFirstRun);
		expect(await fullRows()).toEqual(rows);
	}, 120_000);

	test("--moderation-only reruns keep what an admin already handled", async () => {
		await db.update(reports).set({ status: "resolved" });
		await db.update(contactMessages).set({ handled: true });
		expect(await seedModeration(db, USERS)).toEqual({
			banned: moderationFirstRun.banned,
			hiddenLists: moderationFirstRun.hiddenLists,
			reports: moderationFirstRun.reports,
			contactMessages: moderationFirstRun.contactMessages,
		});
		expect((await db.select({ n: count() }).from(reports).where(eq(reports.status, "open")))[0]?.n).toBe(0);
		expect(
			(await db.select({ n: count() }).from(contactMessages).where(eq(contactMessages.handled, false)))[0]?.n,
		).toBe(0);
	});
});

describe("seed --purge", () => {
	test("deletes seed users with their data and recomputes stats from the remaining ratings", async () => {
		const [item] = await db.select({ id: items.id }).from(items).where(eq(items.slug, "g-0"));
		if (!item) throw new Error("item missing");
		await db.insert(user).values({ id: "real-1", name: "Real", email: "real@example.com", emailVerified: true });
		await db
			.insert(ratings)
			.values({ userId: "real-1", itemId: item.id, c1: 90, c2: 90, c3: 90, c4: 90, c5: 90, c6: 90 });
		await db.insert(lists).values({ userId: "real-1", type: "game", title: "Mine" });

		expect(await purgeSeedUsers(db, 7)).toBe(USERS);
		expect((await db.select({ n: count() }).from(user))[0]?.n).toBe(1);
		const after = await counts();
		expect(after).toEqual({ ratings: 1, statuses: 0, ttb: 0, follows: 0, lists: 1, listItems: 0 });
		expect(await moderation()).toEqual({ banned: 0, hiddenLists: 0, reports: 0, contactMessages: 0 });
		const [rated] = await db.select().from(items).where(eq(items.id, item.id));
		expect([rated?.ratingCount, rated?.score]).toEqual([1, 90]);
		expect(
			await countWhere(
				items,
				sql`${items.id} <> ${item.id} and (${items.ratingCount} <> 0 or ${items.score} is not null)`,
			),
		).toBe(0);
	}, 60_000);
});
