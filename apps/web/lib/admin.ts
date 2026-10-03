import { and, count, desc, eq, inArray, max, sql } from "drizzle-orm";
import { z } from "zod";
import type { Database } from "./db/client";
import { user } from "./db/schema/auth";
import {
	affiliateLinks,
	affiliateStore,
	externalScores,
	type ItemType,
	items,
	itemTranslations,
	sourceRecords,
} from "./db/schema/catalog";
import { contactMessages, itemStatus, listItems, lists, ratings, reports, timeToBeat } from "./db/schema/social";
import { recomputeItemStats } from "./ratings";
import { httpsUrlOrNull } from "./safe-url";

// Data access for /admin. Callers (actions/admin.ts, app/[locale]/admin) check isAdmin() first.

export const AFFILIATE_STORES = affiliateStore.enumValues;
export type AffiliateStore = (typeof AFFILIATE_STORES)[number];

const STORE_DOMAINS: Record<AffiliateStore, readonly string[]> = {
	amazon: [
		"amazon.com",
		"amazon.com.br",
		"amazon.es",
		"amazon.com.mx",
		"amazon.ca",
		"amazon.co.uk",
		"amazon.de",
		"amazon.fr",
		"amazon.it",
		"amzn.to",
	],
	mercadolivre: ["mercadolivre.com.br", "mercadolivre.com", "mercadolibre.com", "mercadolibre.com.mx"],
	steam: ["store.steampowered.com"],
	gog: ["gog.com"],
};

/** https only, no credentials, host equal to (or a subdomain of) one of the store's domains. */
export function isAllowedAffiliateUrl(store: AffiliateStore, value: string): boolean {
	let url: URL;
	try {
		url = new URL(value);
	} catch {
		return false;
	}
	if (url.protocol !== "https:" || url.username || url.password || url.port) return false;
	const host = url.hostname.toLowerCase();
	return STORE_DOMAINS[store].some((d) => host === d || host.endsWith(`.${d}`));
}

export const affiliateInputSchema = z
	.object({
		itemId: z.uuid(),
		store: z.enum(AFFILIATE_STORES),
		url: z.string().trim().max(2048),
		active: z.boolean(),
	})
	.refine((v) => isAllowedAffiliateUrl(v.store, v.url), { path: ["url"], message: "url.domain" });
export type AffiliateInput = z.infer<typeof affiliateInputSchema>;

export async function upsertAffiliateLink(db: Database, input: AffiliateInput): Promise<boolean> {
	const [item] = await db.select({ id: items.id }).from(items).where(eq(items.id, input.itemId)).limit(1);
	if (!item) return false;
	await db
		.insert(affiliateLinks)
		.values(input)
		.onConflictDoUpdate({
			target: [affiliateLinks.itemId, affiliateLinks.store],
			set: { url: input.url, active: input.active },
		});
	return true;
}

export async function deleteAffiliateLink(db: Database, itemId: string, store: AffiliateStore): Promise<void> {
	await db.delete(affiliateLinks).where(and(eq(affiliateLinks.itemId, itemId), eq(affiliateLinks.store, store)));
}

export type MergeResult = "ok" | "notFound" | "same" | "typeMismatch";

/**
 * Moves ratings, statuses, time-to-beat reports, list entries, affiliate links, external scores and source records from `sourceId` to
 * `targetId`, deletes the source and recomputes the target stats — all in one transaction. When a user has a
 * row on both items, the target's row wins (the source row is dropped with the source item).
 */
export function mergeItems(db: Database, sourceId: string, targetId: string): Promise<MergeResult> {
	if (sourceId === targetId) return Promise.resolve("same");
	return db.transaction(async (tx) => {
		const locked = await tx
			.select({ id: items.id, type: items.type })
			.from(items)
			.where(inArray(items.id, [sourceId, targetId]))
			.orderBy(items.id)
			.for("update");
		if (locked.length !== 2) return "notFound";
		if (locked[0]?.type !== locked[1]?.type) return "typeMismatch";
		const target = sql<string>`${targetId}::uuid`.mapWith(String);

		// Raw SQL: drizzle's insert…select demands every column, and `overall` is generated (cannot be inserted).
		await tx.execute(sql`
			insert into ${ratings} (user_id, item_id, c1, c2, c3, c4, c5, c6, created_at, updated_at)
			select user_id, ${targetId}::uuid, c1, c2, c3, c4, c5, c6, created_at, updated_at
			from ${ratings} where item_id = ${sourceId}::uuid
			on conflict do nothing`);
		await tx
			.insert(itemStatus)
			.select(
				tx
					.select({
						userId: itemStatus.userId,
						itemId: target.as("item_id"),
						status: itemStatus.status,
						updatedAt: itemStatus.updatedAt,
					})
					.from(itemStatus)
					.where(eq(itemStatus.itemId, sourceId)),
			)
			.onConflictDoNothing();
		// Only where the user's (merged) target status is `completed`, so the time-to-beat invariant holds.
		await tx.execute(sql`
			insert into ${timeToBeat} (user_id, item_id, main_minutes, extras_minutes, completionist_minutes, created_at, updated_at)
			select t.user_id, ${targetId}::uuid, t.main_minutes, t.extras_minutes, t.completionist_minutes, t.created_at, t.updated_at
			from ${timeToBeat} t
			join ${itemStatus} s on s.user_id = t.user_id and s.item_id = ${targetId}::uuid and s.status = 'completed'
			where t.item_id = ${sourceId}::uuid
			on conflict do nothing`);
		await tx
			.insert(listItems)
			.select(
				tx
					.select({
						listId: listItems.listId,
						itemId: target.as("item_id"),
						position: listItems.position,
						note: listItems.note,
					})
					.from(listItems)
					.where(eq(listItems.itemId, sourceId)),
			)
			.onConflictDoNothing();
		await tx
			.insert(affiliateLinks)
			.select(
				tx
					.select({
						itemId: target.as("item_id"),
						store: affiliateLinks.store,
						url: affiliateLinks.url,
						active: affiliateLinks.active,
					})
					.from(affiliateLinks)
					.where(eq(affiliateLinks.itemId, sourceId)),
			)
			.onConflictDoNothing();
		await tx
			.insert(externalScores)
			.select(
				tx
					.select({
						itemId: target.as("item_id"),
						source: externalScores.source,
						value: externalScores.value,
						url: externalScores.url,
						fetchedAt: externalScores.fetchedAt,
					})
					.from(externalScores)
					.where(eq(externalScores.itemId, sourceId)),
			)
			.onConflictDoNothing();
		await tx.update(sourceRecords).set({ itemId: targetId }).where(eq(sourceRecords.itemId, sourceId));
		await tx.delete(items).where(eq(items.id, sourceId));
		await recomputeItemStats(tx, targetId);
		return "ok";
	});
}

export type IngestionStatus = {
	source: string;
	records: number;
	pending: number;
	lastFetchedAt: string | null;
	lastNormalizedAt: string | null;
};

export async function getIngestionStatus(db: Database): Promise<IngestionStatus[]> {
	const rows = await db
		.select({
			source: sourceRecords.source,
			records: count(),
			pending: sql<number>`count(*) filter (where ${sourceRecords.normalizedAt} is null or ${sourceRecords.normalizedAt} < ${sourceRecords.fetchedAt})::int`,
			lastFetchedAt: max(sourceRecords.fetchedAt),
			lastNormalizedAt: max(sourceRecords.normalizedAt),
		})
		.from(sourceRecords)
		.groupBy(sourceRecords.source)
		.orderBy(sourceRecords.source);
	const iso = (d: Date | string | null) => (d ? new Date(d).toISOString() : null);
	return rows.map((r) => ({
		...r,
		lastFetchedAt: iso(r.lastFetchedAt),
		lastNormalizedAt: iso(r.lastNormalizedAt),
	}));
}

export type AdminReport = {
	id: string;
	targetType: "list" | "user";
	targetId: string;
	targetLabel: string | null;
	targetOwnerId: string | null;
	listHidden: boolean | null;
	reporter: string | null;
	reason: string;
	details: string | null;
	status: "open" | "resolved" | "dismissed";
	createdAt: string;
};

export async function listReports(db: Database, status: "open" | "closed"): Promise<AdminReport[]> {
	const rows = await db
		.select({
			id: reports.id,
			targetType: reports.targetType,
			targetId: reports.targetId,
			reporter: user.username,
			reason: reports.reason,
			details: reports.details,
			status: reports.status,
			createdAt: reports.createdAt,
		})
		.from(reports)
		.leftJoin(user, eq(user.id, reports.reporterId))
		.where(status === "open" ? eq(reports.status, "open") : sql`${reports.status} <> 'open'`)
		.orderBy(desc(reports.createdAt))
		.limit(200);
	const listIds = rows.filter((r) => r.targetType === "list").map((r) => r.targetId);
	const userIds = rows.filter((r) => r.targetType === "user").map((r) => r.targetId);
	const validListIds = listIds.filter((id) => z.uuid().safeParse(id).success);
	const [listRows, userRows] = await Promise.all([
		validListIds.length
			? db
					.select({ id: lists.id, title: lists.title, ownerId: lists.userId, hidden: lists.hidden })
					.from(lists)
					.where(inArray(lists.id, validListIds))
			: [],
		userIds.length
			? db
					.select({ id: user.id, nickname: user.username, banned: user.banned })
					.from(user)
					.where(inArray(user.id, userIds))
			: [],
	]);
	const listMap = new Map(listRows.map((l) => [l.id, l]));
	const userMap = new Map(userRows.map((u) => [u.id, u]));
	return rows.map((r) => {
		const list = r.targetType === "list" ? listMap.get(r.targetId) : undefined;
		const target = r.targetType === "user" ? userMap.get(r.targetId) : undefined;
		return {
			...r,
			targetLabel: list?.title ?? target?.nickname ?? null,
			targetOwnerId: list?.ownerId ?? target?.id ?? null,
			listHidden: list ? list.hidden : null,
			createdAt: new Date(r.createdAt).toISOString(),
		};
	});
}

export async function setReportStatus(db: Database, id: string, status: "resolved" | "dismissed"): Promise<boolean> {
	const rows = await db
		.update(reports)
		.set({ status })
		.where(and(eq(reports.id, id), eq(reports.status, "open")))
		.returning({ id: reports.id });
	return rows.length > 0;
}

export async function setListHidden(db: Database, listId: string, hidden: boolean): Promise<boolean> {
	const rows = await db.update(lists).set({ hidden }).where(eq(lists.id, listId)).returning({ id: lists.id });
	return rows.length > 0;
}

export async function listContactMessages(db: Database, handled: boolean) {
	const rows = await db
		.select()
		.from(contactMessages)
		.where(eq(contactMessages.handled, handled))
		.orderBy(desc(contactMessages.createdAt))
		.limit(200);
	return rows.map((r) => ({ ...r, createdAt: new Date(r.createdAt).toISOString() }));
}

export async function setContactHandled(db: Database, id: string, handled: boolean): Promise<boolean> {
	const rows = await db
		.update(contactMessages)
		.set({ handled })
		.where(eq(contactMessages.id, id))
		.returning({ id: contactMessages.id });
	return rows.length > 0;
}

export type AdminItem = {
	id: string;
	type: ItemType;
	slug: string;
	title: string;
	ratingCount: number;
	score: number | null;
	coverUrl: string | null;
	coverThumbUrl: string | null;
};

const adminItemColumns = {
	id: items.id,
	type: items.type,
	slug: items.slug,
	title: sql<string>`coalesce((select ${itemTranslations.title} from ${itemTranslations} where ${itemTranslations.itemId} = ${items.id} and ${itemTranslations.locale} = 'en'), ${items.slug})`,
	ratingCount: items.ratingCount,
	score: items.score,
	coverUrl: sql<string | null>`${items.meta}->>'coverUrl'`,
	coverThumbUrl: sql<string | null>`${items.meta}->>'coverThumbUrl'`,
};

// Cover URLs come from ingestion-written jsonb: anything but https is dropped.
const safeCovers = (rows: AdminItem[]): AdminItem[] =>
	rows.map((r) => ({ ...r, coverUrl: httpsUrlOrNull(r.coverUrl), coverThumbUrl: httpsUrlOrNull(r.coverThumbUrl) }));

/** By exact id, or slug / English title substring (case-insensitive). */
export async function findItems(db: Database, q: string): Promise<AdminItem[]> {
	const term = q.trim();
	if (!term) return [];
	if (z.uuid().safeParse(term).success) {
		return safeCovers(await db.select(adminItemColumns).from(items).where(eq(items.id, term)));
	}
	const like = `%${term.replace(/[%_\\]/g, "\\$&")}%`;
	const rows = await db
		.select(adminItemColumns)
		.from(items)
		.where(
			sql`${items.slug} ilike ${like} or exists (select 1 from ${itemTranslations} where ${itemTranslations.itemId} = ${items.id} and ${itemTranslations.title} ilike ${like})`,
		)
		.orderBy(desc(items.ratingCount), items.slug)
		.limit(50);
	return safeCovers(rows);
}

export async function getAdminItem(
	db: Database,
	id: string,
): Promise<(AdminItem & { affiliates: { store: AffiliateStore; url: string; active: boolean }[] }) | null> {
	if (!z.uuid().safeParse(id).success) return null;
	const [item] = await db.select(adminItemColumns).from(items).where(eq(items.id, id));
	if (!item) return null;
	const affiliates = await db
		.select({ store: affiliateLinks.store, url: affiliateLinks.url, active: affiliateLinks.active })
		.from(affiliateLinks)
		.where(eq(affiliateLinks.itemId, id))
		.orderBy(affiliateLinks.store);
	return { ...item, affiliates };
}

export async function getAdminOverview(
	db: Database,
): Promise<{ openReports: number; unhandledMessages: number; hiddenLists: number; bannedUsers: number }> {
	const [[r], [m], [l], [u]] = await Promise.all([
		db.select({ n: count() }).from(reports).where(eq(reports.status, "open")),
		db.select({ n: count() }).from(contactMessages).where(eq(contactMessages.handled, false)),
		db.select({ n: count() }).from(lists).where(eq(lists.hidden, true)),
		db.select({ n: count() }).from(user).where(eq(user.banned, true)),
	]);
	return {
		openReports: r?.n ?? 0,
		unhandledMessages: m?.n ?? 0,
		hiddenLists: l?.n ?? 0,
		bannedUsers: u?.n ?? 0,
	};
}
