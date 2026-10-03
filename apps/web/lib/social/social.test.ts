import { afterAll, beforeAll, describe, expect, test } from "bun:test";
import { eq } from "drizzle-orm";
import {
	getIngestionStatus,
	isAllowedAffiliateUrl,
	listReports,
	mergeItems,
	setListHidden,
	upsertAffiliateLink,
} from "@/lib/admin";
import type { Database } from "@/lib/db/client";
import { user } from "@/lib/db/schema/auth";
import { affiliateLinks, items, itemTranslations, sourceRecords } from "@/lib/db/schema/catalog";
import { listItems, lists, ratings, reports } from "@/lib/db/schema/social";
import { createTestDatabase } from "@/lib/db/testing";
import { writeRating } from "@/lib/ratings";
import {
	addListItem,
	createList,
	deleteList,
	getList,
	getListsForItem,
	getUserLists,
	LIST_LIMITS,
	listInputSchema,
	moveListItem,
	removeListItem,
	updateList,
} from "./lists";
import { createReport, reportInputSchema } from "./reports";
import {
	getFeed,
	getItemStatus,
	getProfile,
	getProfileRatings,
	getProfileStats,
	getStatusShelves,
	getTopRaters,
	isFollowing,
	setFollow,
	setItemStatus,
} from "./repository";

let db: Database;
let close: () => Promise<void>;
const item = (n: number) => `01900000-0000-7000-8000-${String(n).padStart(12, "0")}`;
const scores = (v: number) => [v, v, v, v, v, v];

beforeAll(async () => {
	({ db, close } = await createTestDatabase());
	for (let n = 1; n <= 4; n++) {
		await db
			.insert(items)
			.values({ id: item(n), type: n === 4 ? "movie" : "game", slug: `item-${n}`, wikidataQid: `Q${n}` });
		await db.insert(itemTranslations).values({ itemId: item(n), locale: "en", title: `Item ${n}` });
	}
	await db.insert(user).values([
		{ id: "alice", name: "A", email: "a@x.io", emailVerified: true, username: "alice", displayUsername: "Alice" },
		{ id: "bob", name: "B", email: "b@x.io", emailVerified: true, username: "bob" },
		{ id: "carol", name: "C", email: "c@x.io", emailVerified: true, username: "carol", profilePrivate: true },
		{ id: "dave", name: "D", email: "d@x.io", emailVerified: true, username: "dave", banned: true },
	]);
});

afterAll(async () => {
	await close();
});

describe("profiles", () => {
	test("lookup is case-insensitive; banned and unknown are not found; private is flagged", async () => {
		expect(await getProfile(db, "ALICE")).toMatchObject({ id: "alice", nickname: "Alice", isPrivate: false });
		expect(await getProfile(db, "carol")).toMatchObject({ isPrivate: true });
		expect(await getProfile(db, "dave")).toBeNull();
		expect(await getProfile(db, "nobody")).toBeNull();
	});

	test("stats and paginated ratings", async () => {
		await writeRating(db, "alice", { itemId: item(1), scores: scores(90) });
		await writeRating(db, "alice", { itemId: item(2), scores: scores(40) });
		await writeRating(db, "alice", { itemId: item(4), scores: scores(71) });
		const stats = await getProfileStats(db, "alice");
		expect(stats.byType).toEqual({ game: 2, movie: 1, series: 0, book: 0 });
		expect(stats.total).toBe(3);
		expect(stats.averageOverall).toBe(67);
		const best = await getProfileRatings(db, "alice", { sort: "best", page: 1, locale: "en" });
		expect(best.data.map((r) => r.overall)).toEqual([90, 71, 40]);
		expect(best.data[0]?.item).toMatchObject({ title: "Item 1", type: "game", slug: "item-1" });
		expect(best.hasMore).toBe(false);
		const worst = await getProfileRatings(db, "alice", { sort: "worst", page: 1, locale: "en" });
		expect(worst.data[0]?.overall).toBe(40);
	});
});

describe("follows and feed", () => {
	test("follow rules", async () => {
		expect(await setFollow(db, "bob", "alice", true)).toBe("ok");
		expect(await setFollow(db, "bob", "alice", true)).toBe("ok"); // idempotent
		expect(await isFollowing(db, "bob", "alice")).toBe(true);
		expect(await setFollow(db, "bob", "bob", true)).toBe("self");
		expect(await setFollow(db, "bob", "carol", true)).toBe("notFound"); // private
		expect(await setFollow(db, "bob", "dave", true)).toBe("notFound"); // banned
		expect((await getProfileStats(db, "alice")).followers).toBe(1);
		expect((await getProfileStats(db, "bob")).following).toBe(1);
	});

	test("feed shows followed users' ratings newest first with a cursor", async () => {
		const first = await getFeed(db, "bob", { locale: "en", limit: 2 });
		expect(first.data).toHaveLength(2);
		expect(first.data[0]?.nickname).toBe("Alice");
		expect(first.nextCursor).not.toBeNull();
		const second = await getFeed(db, "bob", { locale: "en", limit: 2, cursor: first.nextCursor ?? "" });
		expect(second.data).toHaveLength(1);
		expect(second.nextCursor).toBeNull();
		const all = [...first.data, ...second.data].map((e) => e.item.id);
		expect(new Set(all).size).toBe(3);
		expect(getFeed(db, "bob", { locale: "en", cursor: "garbage" })).rejects.toThrow();
		expect((await getFeed(db, "alice", { locale: "en" })).data).toEqual([]);
	});

	test("top raters exclude self, followed, private and banned users", async () => {
		await writeRating(db, "carol", { itemId: item(1), scores: scores(10) });
		await writeRating(db, "dave", { itemId: item(1), scores: scores(10) });
		expect(await getTopRaters(db, "carol")).toEqual([{ nickname: "Alice", ratings: 3 }]);
		expect(await getTopRaters(db, "bob")).toEqual([]);
		expect(await setFollow(db, "bob", "alice", false)).toBe("ok");
		expect(await isFollowing(db, "bob", "alice")).toBe(false);
	});
});

describe("item status", () => {
	test("set, change, clear and group into shelves", async () => {
		expect(await setItemStatus(db, "bob", item(1), "want")).toBe(true);
		expect(await setItemStatus(db, "bob", item(1), "completed")).toBe(true);
		expect(await setItemStatus(db, "bob", item(2), "dropped")).toBe(true);
		expect(await setItemStatus(db, "bob", "01900000-0000-7000-8000-0000000000ff", "want")).toBe(false);
		expect(await getItemStatus(db, "bob", item(1))).toBe("completed");
		const shelves = await getStatusShelves(db, "bob", "en");
		expect(shelves.completed.map((i) => i.id)).toEqual([item(1)]);
		expect(shelves.dropped.map((i) => i.id)).toEqual([item(2)]);
		expect(shelves.want).toEqual([]);
		await setItemStatus(db, "bob", item(2), null);
		expect(await getItemStatus(db, "bob", item(2))).toBeNull();
	});
});

describe("lists", () => {
	const input = listInputSchema.parse({ title: " Best ", description: "", isPublic: true });

	test("schema limits", () => {
		expect(input).toEqual({ title: "Best", description: null, isPublic: true });
		expect(listInputSchema.safeParse({ ...input, title: "x".repeat(81) }).success).toBe(false);
		expect(listInputSchema.safeParse({ ...input, title: "  " }).success).toBe(false);
		expect(listInputSchema.safeParse({ ...input, description: "x".repeat(501) }).success).toBe(false);
	});

	test("create, add, reorder, remove, visibility, delete", async () => {
		const created = await createList(db, "alice", input);
		if (created === "limit") throw new Error("unexpected limit");
		const id = created.id;
		for (const n of [1, 2, 3]) expect(await addListItem(db, "alice", id, item(n))).toBe("ok");
		expect(await addListItem(db, "alice", id, item(1))).toBe("ok"); // duplicate is a no-op
		expect(await addListItem(db, "bob", id, item(4))).toBe("notFound"); // not the owner
		expect(await moveListItem(db, "alice", id, item(3), "up")).toBe(true);
		expect(await moveListItem(db, "alice", id, item(1), "up")).toBe(true); // already first
		const list = await getList(db, id, "bob", "en");
		expect(list?.items.map((i) => i.id)).toEqual([item(1), item(3), item(2)]);
		expect(list?.itemCount).toBe(3);
		expect(await removeListItem(db, "alice", id, item(3))).toBe(true);
		expect((await getListsForItem(db, "alice", item(2)))[0]).toMatchObject({ id, contains: true });
		expect((await getListsForItem(db, "alice", item(3)))[0]).toMatchObject({ id, contains: false });

		expect(await updateList(db, "bob", id, { ...input, isPublic: false })).toBe(false);
		expect(await updateList(db, "alice", id, { ...input, isPublic: false })).toBe(true);
		expect(await getList(db, id, "bob", "en")).toBeNull();
		expect(await getList(db, id, null, "en")).toBeNull();
		expect(await getList(db, id, "alice", "en")).not.toBeNull();
		expect(await getUserLists(db, "alice", false)).toEqual([]);
		expect(await getUserLists(db, "alice", true)).toHaveLength(1);

		await updateList(db, "alice", id, input);
		expect(await setListHidden(db, id, true)).toBe(true);
		expect(await getList(db, id, "bob", "en")).toBeNull();
		expect((await getList(db, id, "alice", "en"))?.hidden).toBe(true);
		await setListHidden(db, id, false);

		expect(await deleteList(db, "bob", id)).toBe(false);
		expect(await deleteList(db, "alice", id)).toBe(true);
		expect(await getList(db, id, "alice", "en")).toBeNull();
	});

	test("lists of a private profile are hidden from others", async () => {
		const created = await createList(db, "carol", input);
		if (created === "limit") throw new Error("unexpected limit");
		expect(await getList(db, created.id, "bob", "en")).toBeNull();
		expect(await getList(db, created.id, "carol", "en")).not.toBeNull();
	});

	test("caps the number of lists per user", async () => {
		await db
			.insert(lists)
			.values(Array.from({ length: LIST_LIMITS.lists }, (_, i) => ({ userId: "bob", title: `L${i}` })));
		expect(await createList(db, "bob", input)).toBe("limit");
	});
});

describe("reports", () => {
	test("validates, dedupes open reports and rejects self-reports", async () => {
		const parse = (v: unknown) => reportInputSchema.parse(v);
		expect(
			reportInputSchema.safeParse({ target: { type: "list", id: "x" }, reason: "spam", details: "" }).success,
		).toBe(false);
		expect(
			reportInputSchema.safeParse({ target: { type: "user", nickname: "a" }, reason: "nope", details: "" })
				.success,
		).toBe(false);
		const report = parse({ target: { type: "user", nickname: "Alice" }, reason: "spam", details: "  " });
		expect(await createReport(db, "bob", report)).toBe("ok");
		expect(await createReport(db, "bob", report)).toBe("ok");
		expect(await db.select().from(reports)).toHaveLength(1);
		expect(await createReport(db, "alice", report)).toBe("self");
		expect(
			await createReport(
				db,
				"bob",
				parse({ target: { type: "user", nickname: "ghost" }, reason: "spam", details: "" }),
			),
		).toBe("notFound");

		const [list] = await db.insert(lists).values({ userId: "alice", title: "Spammy" }).returning();
		if (!list) throw new Error("no list");
		const listReport = parse({ target: { type: "list", id: list.id }, reason: "offensive", details: "rude title" });
		expect(await createReport(db, "bob", listReport)).toBe("ok");
		const open = await listReports(db, "open");
		expect(open).toHaveLength(2);
		expect(open.find((r) => r.targetType === "list")).toMatchObject({
			targetLabel: "Spammy",
			listHidden: false,
			details: "rude title",
			reporter: "bob",
		});
		await db.update(reports).set({ status: "resolved" });
		expect(await createReport(db, "bob", report)).toBe("ok"); // a closed report does not block a new one
		expect(await listReports(db, "closed")).toHaveLength(2);
	});
});

describe("admin", () => {
	test("affiliate URLs must be https on the store's domains", () => {
		expect(isAllowedAffiliateUrl("amazon", "https://www.amazon.com.br/dp/B0?tag=x")).toBe(true);
		expect(isAllowedAffiliateUrl("amazon", "https://amzn.to/abc")).toBe(true);
		expect(isAllowedAffiliateUrl("amazon", "http://www.amazon.com/dp/B0")).toBe(false);
		expect(isAllowedAffiliateUrl("amazon", "https://amazon.com.evil.io/")).toBe(false);
		expect(isAllowedAffiliateUrl("amazon", "https://user:pw@amazon.com/")).toBe(false);
		expect(isAllowedAffiliateUrl("steam", "https://store.steampowered.com/app/1")).toBe(true);
		expect(isAllowedAffiliateUrl("steam", "https://steampowered.com.evil/")).toBe(false);
		expect(isAllowedAffiliateUrl("gog", "https://www.gog.com/en/game/x")).toBe(true);
		expect(isAllowedAffiliateUrl("mercadolivre", "https://produto.mercadolivre.com.br/MLB-1")).toBe(true);
		expect(isAllowedAffiliateUrl("gog", "javascript:alert(1)")).toBe(false);
	});

	test("merge moves everything to the target and recomputes its stats", async () => {
		const [source, target] = [item(2), item(3)];
		await writeRating(db, "bob", { itemId: source, scores: scores(20) }); // only on source → moves
		await writeRating(db, "carol", { itemId: target, scores: scores(80) });
		await writeRating(db, "carol", { itemId: source, scores: scores(0) }); // on both → target wins
		await setItemStatus(db, "carol", source, "want");
		await upsertAffiliateLink(db, {
			itemId: source,
			store: "steam",
			url: "https://store.steampowered.com/app/1",
			active: true,
		});
		await db
			.insert(sourceRecords)
			.values({ source: "steam", externalId: "1", payload: {}, hash: "h", itemId: source });
		const [list] = await db.insert(lists).values({ userId: "carol", title: "Mine" }).returning();
		if (!list) throw new Error("no list");
		await db.insert(listItems).values({ listId: list.id, itemId: source, position: 1 });

		expect(await mergeItems(db, source, source)).toBe("same");
		expect(await mergeItems(db, source, item(4))).toBe("typeMismatch");
		expect(await mergeItems(db, source, "01900000-0000-7000-8000-0000000000ff")).toBe("notFound");
		expect(await mergeItems(db, source, target)).toBe("ok");

		expect(await db.select().from(items).where(eq(items.id, source))).toEqual([]);
		const moved = await db.select().from(ratings).where(eq(ratings.itemId, target));
		// alice (40 → from source), bob (20), carol (80, kept)
		expect(moved.map((r) => [r.userId, r.overall]).sort()).toEqual([
			["alice", 40],
			["bob", 20],
			["carol", 80],
		]);
		const [stats] = await db.select().from(items).where(eq(items.id, target));
		expect(stats?.ratingCount).toBe(3);
		expect(stats?.score).toBe(47); // (40 + 20 + 80) / 3 = 46.67
		expect(await getItemStatus(db, "carol", target)).toBe("want");
		expect(await db.select().from(affiliateLinks).where(eq(affiliateLinks.itemId, target))).toHaveLength(1);
		expect((await db.select().from(sourceRecords))[0]?.itemId).toBe(target);
		expect(await db.select().from(listItems).where(eq(listItems.listId, list.id))).toMatchObject([
			{ itemId: target },
		]);
	});

	test("ingestion status per source", async () => {
		await db.insert(sourceRecords).values({ source: "steam", externalId: "2", payload: {}, hash: "h" });
		const [steam] = await getIngestionStatus(db);
		expect(steam).toMatchObject({ source: "steam", records: 2, pending: 2, lastNormalizedAt: null });
		expect(steam?.lastFetchedAt).toBeString();
	});
});
