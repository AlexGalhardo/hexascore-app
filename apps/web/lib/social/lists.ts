import { and, asc, count, desc, eq, gt, lt, max, sql } from "drizzle-orm";
import { z } from "zod";
import type { Locale } from "@/i18n/routing";
import type { Database } from "@/lib/db/client";
import { user } from "@/lib/db/schema/auth";
import { items } from "@/lib/db/schema/catalog";
import { listItems, lists } from "@/lib/db/schema/social";
import { LIST_LIMITS } from "./constants";
import { displayName, type ItemRef, itemRefColumns, publicUser, titleJoin } from "./repository";

export { LIST_LIMITS };

export const listInputSchema = z.object({
	title: z.string().trim().min(1).max(LIST_LIMITS.title),
	description: z
		.string()
		.trim()
		.max(LIST_LIMITS.description)
		.transform((v) => v || null),
	isPublic: z.boolean(),
});
export type ListInput = z.infer<typeof listInputSchema>;

export type ListSummary = {
	id: string;
	title: string;
	description: string | null;
	isPublic: boolean;
	hidden: boolean;
	itemCount: number;
	updatedAt: string;
};
export type ListDetail = ListSummary & { ownerId: string; ownerNickname: string; items: ItemRef[] };

const summaryColumns = {
	id: lists.id,
	title: lists.title,
	description: lists.description,
	isPublic: lists.isPublic,
	hidden: lists.hidden,
	updatedAt: lists.updatedAt,
	itemCount: sql<number>`(select count(*)::int from ${listItems} where ${listItems.listId} = ${lists.id})`,
};

const toSummary = (r: Omit<ListSummary, "updatedAt"> & { updatedAt: Date }): ListSummary => ({
	...r,
	updatedAt: new Date(r.updatedAt).toISOString(),
});

/** Owner sees every list (moderated ones flagged); others only public, non-hidden lists. */
export async function getUserLists(db: Database, ownerId: string, isOwner: boolean): Promise<ListSummary[]> {
	const rows = await db
		.select(summaryColumns)
		.from(lists)
		.where(
			and(
				eq(lists.userId, ownerId),
				isOwner ? undefined : and(eq(lists.isPublic, true), eq(lists.hidden, false)),
			),
		)
		.orderBy(desc(lists.updatedAt));
	return rows.map(toSummary);
}

/**
 * A list page is visible to its owner, or to anyone when it is public, not hidden by moderation and its owner's
 * profile is public. Anything else is indistinguishable from a missing list (404).
 */
export async function getList(
	db: Database,
	listId: string,
	viewerId: string | null,
	locale: Locale,
): Promise<ListDetail | null> {
	if (!z.uuid().safeParse(listId).success) return null;
	const [row] = await db
		.select({
			...summaryColumns,
			ownerId: lists.userId,
			ownerNickname: displayName,
			ownerPublic: sql<boolean>`${publicUser}`,
		})
		.from(lists)
		.innerJoin(user, eq(user.id, lists.userId))
		.where(eq(lists.id, listId))
		.limit(1);
	if (!row) return null;
	const isOwner = viewerId === row.ownerId;
	if (!isOwner && (!row.isPublic || row.hidden || !row.ownerPublic)) return null;
	const j = titleJoin(locale);
	const entries = await db
		.select(itemRefColumns)
		.from(listItems)
		.innerJoin(items, eq(items.id, listItems.itemId))
		.leftJoin(...j.tl)
		.leftJoin(...j.ten)
		.where(eq(listItems.listId, listId))
		.orderBy(asc(listItems.position));
	const { ownerPublic: _, ownerId, ownerNickname, ...rest } = row;
	return { ...toSummary(rest), ownerId, ownerNickname, items: entries };
}

/** The viewer's lists with a flag for whether `itemId` is already in each (item page menu). */
export async function getListsForItem(
	db: Database,
	userId: string,
	itemId: string,
): Promise<{ id: string; title: string; contains: boolean }[]> {
	return db
		.select({
			id: lists.id,
			title: lists.title,
			contains: sql<boolean>`exists (select 1 from ${listItems} where ${listItems.listId} = ${lists.id} and ${listItems.itemId} = ${itemId})`,
		})
		.from(lists)
		.where(eq(lists.userId, userId))
		.orderBy(desc(lists.updatedAt));
}

export type ListWriteResult = "ok" | "notFound" | "limit";

export async function createList(db: Database, userId: string, input: ListInput): Promise<{ id: string } | "limit"> {
	return db.transaction(async (tx) => {
		// Serializes concurrent creates of the same user so the cap cannot be raced past.
		await tx.execute(sql`select pg_advisory_xact_lock(hashtext(${`lists:${userId}`}))`);
		const [existing] = await tx.select({ n: count() }).from(lists).where(eq(lists.userId, userId));
		if ((existing?.n ?? 0) >= LIST_LIMITS.lists) return "limit";
		const [row] = await tx
			.insert(lists)
			.values({ userId, ...input })
			.returning({ id: lists.id });
		if (!row) throw new Error("list insert returned no row");
		return { id: row.id };
	});
}

const owned = (userId: string, listId: string) => and(eq(lists.id, listId), eq(lists.userId, userId));

export async function updateList(db: Database, userId: string, listId: string, input: ListInput): Promise<boolean> {
	const rows = await db.update(lists).set(input).where(owned(userId, listId)).returning({ id: lists.id });
	return rows.length > 0;
}

export async function deleteList(db: Database, userId: string, listId: string): Promise<boolean> {
	const rows = await db.delete(lists).where(owned(userId, listId)).returning({ id: lists.id });
	return rows.length > 0;
}

type Tx = Parameters<Parameters<Database["transaction"]>[0]>[0];

// Locks the list row so concurrent edits of the same list (append, reorder) serialize.
async function lockOwnedList(tx: Tx, userId: string, listId: string): Promise<boolean> {
	const rows = await tx.select({ id: lists.id }).from(lists).where(owned(userId, listId)).for("update");
	return rows.length > 0;
}

const touch = (tx: Tx, listId: string) => tx.update(lists).set({ updatedAt: new Date() }).where(eq(lists.id, listId));

export function addListItem(db: Database, userId: string, listId: string, itemId: string): Promise<ListWriteResult> {
	return db.transaction(async (tx) => {
		if (!(await lockOwnedList(tx, userId, listId))) return "notFound";
		const [item] = await tx.select({ id: items.id }).from(items).where(eq(items.id, itemId)).limit(1);
		if (!item) return "notFound";
		const [stats] = await tx
			.select({ n: count(), last: max(listItems.position) })
			.from(listItems)
			.where(eq(listItems.listId, listId));
		if ((stats?.n ?? 0) >= LIST_LIMITS.items) return "limit";
		await tx
			.insert(listItems)
			.values({ listId, itemId, position: (stats?.last ?? 0) + 1 })
			.onConflictDoNothing();
		await touch(tx, listId);
		return "ok";
	});
}

export function removeListItem(db: Database, userId: string, listId: string, itemId: string): Promise<boolean> {
	return db.transaction(async (tx) => {
		if (!(await lockOwnedList(tx, userId, listId))) return false;
		await tx.delete(listItems).where(and(eq(listItems.listId, listId), eq(listItems.itemId, itemId)));
		await touch(tx, listId);
		return true;
	});
}

/** Swaps the item with its neighbour; moving past either end is a no-op. */
export function moveListItem(
	db: Database,
	userId: string,
	listId: string,
	itemId: string,
	direction: "up" | "down",
): Promise<boolean> {
	return db.transaction(async (tx) => {
		if (!(await lockOwnedList(tx, userId, listId))) return false;
		const [current] = await tx
			.select({ position: listItems.position })
			.from(listItems)
			.where(and(eq(listItems.listId, listId), eq(listItems.itemId, itemId)));
		if (!current) return false;
		const [neighbour] = await tx
			.select({ itemId: listItems.itemId, position: listItems.position })
			.from(listItems)
			.where(
				and(
					eq(listItems.listId, listId),
					direction === "up"
						? lt(listItems.position, current.position)
						: gt(listItems.position, current.position),
				),
			)
			.orderBy(direction === "up" ? desc(listItems.position) : asc(listItems.position))
			.limit(1);
		if (!neighbour) return true;
		await tx
			.update(listItems)
			.set({ position: neighbour.position })
			.where(and(eq(listItems.listId, listId), eq(listItems.itemId, itemId)));
		await tx
			.update(listItems)
			.set({ position: current.position })
			.where(and(eq(listItems.listId, listId), eq(listItems.itemId, neighbour.itemId)));
		await touch(tx, listId);
		return true;
	});
}
