import { afterAll, beforeAll, describe, expect, test } from "bun:test";
import { count, eq, like, max, sql } from "drizzle-orm";
import type { Database } from "./client";
import { user } from "./schema/auth";
import { type ItemType, items } from "./schema/catalog";
import { follows, itemStatus, listItems, lists, ratings, timeToBeat } from "./schema/social";
import { purgeSeedUsers, seedExistingItems } from "./seed";
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

	test("a second run is idempotent", async () => {
		const rows = await fullRows();
		await seedExistingItems(db, OPTS);
		expect(await counts()).toEqual(firstRun);
		expect(await fullRows()).toEqual(rows);
	}, 120_000);
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
