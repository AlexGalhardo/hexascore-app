import { and, count, desc, eq, isNotNull, isNull, notInArray, type SQL, sql } from "drizzle-orm";
import { alias } from "drizzle-orm/pg-core";
import { z } from "zod";
import type { Locale } from "@/i18n/routing";
import { decodeCursor, encodeCursor, InvalidCursorError } from "@/lib/catalog/cursor";
import type { Database } from "@/lib/db/client";
import { user } from "@/lib/db/schema/auth";
import { ITEM_TYPES, type ItemType, items, itemTranslations } from "@/lib/db/schema/catalog";
import { follows, itemStatus, type itemStatusValue, ratings, timeToBeat } from "@/lib/db/schema/social";
import { httpsUrlOrNull } from "@/lib/safe-url";
import { ITEM_STATUSES, type ItemStatus } from "./constants";

// Pure data access for profiles, follows, feed and item status (same contract as lib/catalog/repository.ts:
// `fn(db, …)` returning JSON-only DTOs). Pages call these directly — this data is per-user and not cached.

type SameStatuses = (typeof itemStatusValue.enumValues)[number] extends ItemStatus
	? ItemStatus extends (typeof itemStatusValue.enumValues)[number]
		? true
		: never
	: never;
const _statusesMatch: SameStatuses = true;

export { ITEM_STATUSES, type ItemStatus };
export const itemStatusSchema = z.enum(ITEM_STATUSES);

export type ItemRef = {
	id: string;
	type: ItemType;
	slug: string;
	title: string;
	coverUrl: string | null;
	coverThumbUrl: string | null;
};

const tl = alias(itemTranslations, "tl");
const ten = alias(itemTranslations, "ten");

export const itemRefColumns = {
	id: items.id,
	type: items.type,
	slug: items.slug,
	title: sql<string>`coalesce(${tl.title}, ${ten.title}, ${items.slug})`,
	// Only the two cover keys, not the whole meta jsonb (genres, trailers…): these rows feed light listings.
	coverUrl: sql<string | null>`${items.meta}->>'coverUrl'`,
	coverThumbUrl: sql<string | null>`${items.meta}->>'coverThumbUrl'`,
};

/** jsonb is written by the ingestion, so cover URLs are untrusted: anything but https is dropped. */
export function toItemRef(r: ItemRef): ItemRef {
	return {
		id: r.id,
		type: r.type,
		slug: r.slug,
		title: r.title,
		coverUrl: httpsUrlOrNull(r.coverUrl),
		coverThumbUrl: httpsUrlOrNull(r.coverThumbUrl),
	};
}
export const titleJoin = (locale: Locale) => ({
	tl: [tl, and(eq(tl.itemId, items.id), eq(tl.locale, locale))] as const,
	ten: [ten, and(eq(ten.itemId, items.id), eq(ten.locale, "en"))] as const,
});

/** Users whose profile and activity others may see. */
export const publicUser: SQL = and(
	eq(user.profilePrivate, false),
	isNotNull(user.username),
	sql`coalesce(${user.banned}, false) = false`,
	isNull(user.deletionScheduledAt),
) as SQL;

export const displayName = sql<string>`coalesce(${user.displayUsername}, ${user.username})`;

// ---------------------------------------------------------------------------------------------------------------
// Profiles

export type Profile = { id: string; nickname: string; memberSince: string; isPrivate: boolean };

/** Banned, anonymized and unknown users are all "not found"; private ones are found but flagged. */
export async function getProfile(db: Database, nickname: string): Promise<Profile | null> {
	const [row] = await db
		.select({
			id: user.id,
			nickname: displayName,
			createdAt: user.createdAt,
			profilePrivate: user.profilePrivate,
			deletionScheduledAt: user.deletionScheduledAt,
		})
		.from(user)
		.where(and(eq(user.username, nickname.toLowerCase()), sql`coalesce(${user.banned}, false) = false`))
		.limit(1);
	if (!row) return null;
	return {
		id: row.id,
		nickname: row.nickname,
		memberSince: new Date(row.createdAt).toISOString(),
		isPrivate: row.profilePrivate || row.deletionScheduledAt !== null,
	};
}

export type ProfileStats = {
	byType: Record<ItemType, number>;
	total: number;
	averageOverall: number | null;
	followers: number;
	following: number;
};

export async function getProfileStats(db: Database, userId: string): Promise<ProfileStats> {
	const [byType, [avgRow], [followers], [following]] = await Promise.all([
		db
			.select({ type: items.type, n: count() })
			.from(ratings)
			.innerJoin(items, eq(items.id, ratings.itemId))
			.where(eq(ratings.userId, userId))
			.groupBy(items.type),
		db
			.select({ avg: sql<string | null>`avg(${ratings.overall})` })
			.from(ratings)
			.where(eq(ratings.userId, userId)),
		db
			.select({ n: count() })
			.from(follows)
			.innerJoin(user, eq(user.id, follows.followerId))
			.where(and(eq(follows.followeeId, userId), publicUser)),
		db
			.select({ n: count() })
			.from(follows)
			.innerJoin(user, eq(user.id, follows.followeeId))
			.where(and(eq(follows.followerId, userId), publicUser)),
	]);
	const counts = Object.fromEntries(ITEM_TYPES.map((t) => [t, 0])) as Record<ItemType, number>;
	for (const row of byType) counts[row.type] = row.n;
	return {
		byType: counts,
		total: byType.reduce((sum, r) => sum + r.n, 0),
		averageOverall: avgRow?.avg ? Math.round(Number(avgRow.avg)) : null,
		followers: followers?.n ?? 0,
		following: following?.n ?? 0,
	};
}

export const PROFILE_SORTS = ["recent", "best", "worst"] as const;
export type ProfileSort = (typeof PROFILE_SORTS)[number];
export const PROFILE_PAGE_SIZE = 20;

export type ProfileRating = { item: ItemRef; overall: number; categories: number[]; ratedAt: string };

export async function getProfileRatings(
	db: Database,
	userId: string,
	options: { sort: ProfileSort; page: number; locale: Locale; type?: ItemType; limit?: number },
): Promise<{ data: ProfileRating[]; hasMore: boolean }> {
	const size = options.limit ?? PROFILE_PAGE_SIZE;
	const order = {
		recent: [desc(ratings.createdAt)],
		best: [desc(ratings.overall), desc(ratings.createdAt)],
		worst: [ratings.overall, desc(ratings.createdAt)],
	}[options.sort];
	const j = titleJoin(options.locale);
	const rows = await db
		.select({
			...itemRefColumns,
			overall: ratings.overall,
			c1: ratings.c1,
			c2: ratings.c2,
			c3: ratings.c3,
			c4: ratings.c4,
			c5: ratings.c5,
			c6: ratings.c6,
			createdAt: ratings.createdAt,
		})
		.from(ratings)
		.innerJoin(items, eq(items.id, ratings.itemId))
		.leftJoin(...j.tl)
		.leftJoin(...j.ten)
		.where(and(eq(ratings.userId, userId), options.type ? eq(items.type, options.type) : undefined))
		.orderBy(...order, ratings.itemId)
		.limit(size + 1)
		.offset((Math.max(options.page, 1) - 1) * size);
	return {
		data: rows.slice(0, size).map((r) => ({
			item: toItemRef(r),
			overall: r.overall ?? 0,
			categories: [r.c1, r.c2, r.c3, r.c4, r.c5, r.c6],
			ratedAt: new Date(r.createdAt).toISOString(),
		})),
		hasMore: rows.length > size,
	};
}

export async function getStatusShelves(
	db: Database,
	userId: string,
	locale: Locale,
): Promise<Record<ItemStatus, ItemRef[]>> {
	const j = titleJoin(locale);
	const rows = await db
		.select({ ...itemRefColumns, status: itemStatus.status })
		.from(itemStatus)
		.innerJoin(items, eq(items.id, itemStatus.itemId))
		.leftJoin(...j.tl)
		.leftJoin(...j.ten)
		.where(eq(itemStatus.userId, userId))
		.orderBy(desc(itemStatus.updatedAt))
		.limit(400);
	const shelves = Object.fromEntries(ITEM_STATUSES.map((s) => [s, [] as ItemRef[]])) as Record<ItemStatus, ItemRef[]>;
	for (const r of rows) shelves[r.status].push(toItemRef(r));
	return shelves;
}

// ---------------------------------------------------------------------------------------------------------------
// Follows

export async function isFollowing(db: Database, followerId: string, followeeId: string): Promise<boolean> {
	const [row] = await db
		.select({ id: follows.followeeId })
		.from(follows)
		.where(and(eq(follows.followerId, followerId), eq(follows.followeeId, followeeId)))
		.limit(1);
	return Boolean(row);
}

export type FollowResult = "ok" | "notFound" | "self";

/** Following needs a public profile; unfollowing always works (the target may have gone private). */
export async function setFollow(
	db: Database,
	followerId: string,
	nickname: string,
	follow: boolean,
): Promise<FollowResult> {
	const [target] = await db
		.select({ id: user.id })
		.from(user)
		.where(and(eq(user.username, nickname.toLowerCase()), follow ? publicUser : undefined))
		.limit(1);
	if (!target) return "notFound";
	if (target.id === followerId) return "self";
	if (follow) {
		await db.insert(follows).values({ followerId, followeeId: target.id }).onConflictDoNothing();
	} else {
		await db.delete(follows).where(and(eq(follows.followerId, followerId), eq(follows.followeeId, target.id)));
	}
	return "ok";
}

// ---------------------------------------------------------------------------------------------------------------
// Feed

export type FeedEntry = { nickname: string; item: ItemRef; overall: number; ratedAt: string };

const feedCursorSchema = z.object({ t: z.string().max(64), u: z.string().max(128), i: z.uuid() });

export async function getFeed(
	db: Database,
	userId: string,
	options: { cursor?: string; locale: Locale; limit?: number },
): Promise<{ data: FeedEntry[]; nextCursor: string | null }> {
	const limit = Math.min(Math.max(options.limit ?? 20, 1), 50);
	const cursor = options.cursor ? decodeCursor(options.cursor, feedCursorSchema) : null;
	if (options.cursor && !cursor) throw new InvalidCursorError();
	const j = titleJoin(options.locale);
	const rows = await db
		.select({
			...itemRefColumns,
			nickname: displayName,
			overall: ratings.overall,
			createdAt: ratings.createdAt,
			cursorTime: sql<string>`${ratings.createdAt}::text`,
			userId: ratings.userId,
		})
		.from(follows)
		.innerJoin(ratings, eq(ratings.userId, follows.followeeId))
		.innerJoin(user, eq(user.id, follows.followeeId))
		.innerJoin(items, eq(items.id, ratings.itemId))
		.leftJoin(...j.tl)
		.leftJoin(...j.ten)
		.where(
			and(
				eq(follows.followerId, userId),
				publicUser,
				cursor
					? sql`(${ratings.createdAt}, ${ratings.userId}, ${ratings.itemId}) < (${cursor.t}::timestamptz, ${cursor.u}, ${cursor.i}::uuid)`
					: undefined,
			),
		)
		.orderBy(desc(ratings.createdAt), desc(ratings.userId), desc(ratings.itemId))
		.limit(limit + 1);
	const page = rows.slice(0, limit);
	const last = page.at(-1);
	return {
		data: page.map((r) => ({
			nickname: r.nickname,
			item: toItemRef(r),
			overall: r.overall ?? 0,
			ratedAt: new Date(r.createdAt).toISOString(),
		})),
		nextCursor:
			rows.length > limit && last ? encodeCursor({ t: last.cursorTime, u: last.userId, i: last.id }) : null,
	};
}

export type Rater = { nickname: string; ratings: number };

/** Most active public raters the viewer does not follow yet (feed empty state). */
export async function getTopRaters(db: Database, viewerId: string, limit = 8): Promise<Rater[]> {
	const followed = db.select({ id: follows.followeeId }).from(follows).where(eq(follows.followerId, viewerId));
	const rows = await db
		.select({ nickname: displayName, n: count() })
		.from(ratings)
		.innerJoin(user, eq(user.id, ratings.userId))
		.where(and(publicUser, sql`${user.id} <> ${viewerId}`, notInArray(user.id, followed)))
		.groupBy(user.id, user.username, user.displayUsername)
		.orderBy(desc(count()), user.username)
		.limit(limit);
	return rows.map((r) => ({ nickname: r.nickname, ratings: r.n }));
}

// ---------------------------------------------------------------------------------------------------------------
// Item status

export async function getItemStatus(db: Database, userId: string, itemId: string): Promise<ItemStatus | null> {
	const [row] = await db
		.select({ status: itemStatus.status })
		.from(itemStatus)
		.where(and(eq(itemStatus.userId, userId), eq(itemStatus.itemId, itemId)))
		.limit(1);
	return row?.status ?? null;
}

/**
 * `null` clears the status. Leaving `completed` also deletes the user's time-to-beat report (only completers
 * report), in the same transaction: "cleared" tells the caller to refresh the item's cached medians.
 */
export function setItemStatus(
	db: Database,
	userId: string,
	itemId: string,
	status: ItemStatus | null,
): Promise<"ok" | "cleared" | "notFound"> {
	return db.transaction(async (tx) => {
		const [exists] = await tx.select({ id: items.id }).from(items).where(eq(items.id, itemId)).limit(1);
		if (!exists) return "notFound";
		if (status === null) {
			await tx.delete(itemStatus).where(and(eq(itemStatus.userId, userId), eq(itemStatus.itemId, itemId)));
		} else {
			await tx
				.insert(itemStatus)
				.values({ userId, itemId, status })
				.onConflictDoUpdate({
					target: [itemStatus.userId, itemStatus.itemId],
					set: { status, updatedAt: new Date() },
				});
		}
		if (status === "completed") return "ok";
		const cleared = await tx
			.delete(timeToBeat)
			.where(and(eq(timeToBeat.userId, userId), eq(timeToBeat.itemId, itemId)))
			.returning({ itemId: timeToBeat.itemId });
		return cleared.length > 0 ? "cleared" : "ok";
	});
}
