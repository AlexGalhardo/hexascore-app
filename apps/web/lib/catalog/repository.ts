import { and, asc, desc, eq, gte, isNull, lte, or, type SQL, sql } from "drizzle-orm";
import { alias } from "drizzle-orm/pg-core";
import { z } from "zod";
import type { Locale } from "@/i18n/routing";
import { MIN_RATINGS_FOR_SCORE } from "@/lib/categories";
import type { Database } from "@/lib/db/client";
import { user } from "@/lib/db/schema/auth";
import { affiliateLinks, externalScores, type ItemType, items, itemTranslations } from "@/lib/db/schema/catalog";
import { ratings } from "@/lib/db/schema/social";
import { httpsUrlOrNull } from "@/lib/safe-url";
import { getTimeToBeat } from "@/lib/time-to-beat";
import { decodeCursor, encodeCursor, InvalidCursorError } from "./cursor";
import { parseExternalIds, parseMeta } from "./jsonb";
import {
	type AffiliateLink,
	type ExternalScore,
	type ItemDetail,
	type ItemSummary,
	type ListItemsParams,
	listItemsParamsSchema,
	type Page,
	type PublicRating,
	type Sort,
} from "./types";

// Every function takes the database as its first argument: pages use the cached wrappers in ./cached.ts, tests
// pass an in-memory PGlite, and the public API (0.4) reuses these as-is.

const tl = alias(itemTranslations, "tl");
const ten = alias(itemTranslations, "ten");

const visible = gte(items.ratingCount, MIN_RATINGS_FOR_SCORE);
const scoreColumns = [items.scoreC1, items.scoreC2, items.scoreC3, items.scoreC4, items.scoreC5, items.scoreC6];

const summaryColumns = {
	id: items.id,
	type: items.type,
	slug: items.slug,
	releaseDate: items.releaseDate,
	ratingCount: items.ratingCount,
	score: items.score,
	c1: items.scoreC1,
	c2: items.scoreC2,
	c3: items.scoreC3,
	c4: items.scoreC4,
	c5: items.scoreC5,
	c6: items.scoreC6,
	meta: items.meta,
	titleLocale: tl.title,
	titleEn: ten.title,
};

type SummaryRow = {
	id: string;
	type: ItemType;
	slug: string;
	releaseDate: string | null;
	ratingCount: number;
	score: number | null;
	c1: number | null;
	c2: number | null;
	c3: number | null;
	c4: number | null;
	c5: number | null;
	c6: number | null;
	meta: unknown;
	titleLocale: string | null;
	titleEn: string | null;
};

function toSummary(row: SummaryRow): ItemSummary {
	const meta = parseMeta(row.meta, row.id);
	const shown = row.ratingCount >= MIN_RATINGS_FOR_SCORE;
	return {
		id: row.id,
		type: row.type,
		slug: row.slug,
		title: row.titleLocale ?? row.titleEn ?? row.slug,
		releaseDate: row.releaseDate,
		year: row.releaseDate ? Number(row.releaseDate.slice(0, 4)) : null,
		score: shown ? row.score : null,
		categories: shown ? [row.c1, row.c2, row.c3, row.c4, row.c5, row.c6] : [null, null, null, null, null, null],
		ratingCount: row.ratingCount,
		genres: meta.genres ?? [],
		platforms: meta.platforms ?? [],
		coverUrl: meta.coverUrl ?? null,
		coverThumbUrl: meta.coverThumbUrl ?? null,
	};
}

function withTranslations(locale: Locale) {
	return {
		tl: and(eq(tl.itemId, items.id), eq(tl.locale, locale)),
		ten: and(eq(ten.itemId, items.id), eq(ten.locale, "en")),
	};
}

// ---------------------------------------------------------------------------------------------------------------
// listItems — keyset pagination over (sort expression, id)

type Direction = "asc" | "desc";
type SortField = "score" | "release_date" | "rating_count" | "popularity";

const sortCursorSchema = z.object({
	s: z.string(),
	v: z.string().max(64),
	id: z.uuid(),
	d: z.enum(["next", "prev"]),
});

// Nulls (no visible score, unknown date) always sort last, whichever the direction.
function sortExpression(field: SortField, dir: Direction): { expr: SQL; cast: string } {
	switch (field) {
		case "score":
			return {
				expr: sql`coalesce(case when ${visible} then ${items.score} end, ${sql.raw(dir === "asc" ? "1000" : "-1")})`,
				cast: "int",
			};
		case "release_date":
			return {
				expr: sql`coalesce(${items.releaseDate}, ${sql.raw(dir === "asc" ? "'9999-12-31'" : "'0001-01-01'")}::date)`,
				cast: "date",
			};
		case "rating_count":
			return { expr: sql`${items.ratingCount}`, cast: "int" };
		case "popularity":
			return { expr: sql`${items.popularity}`, cast: "real" };
	}
}

function parseSort(sort: Sort): { field: SortField; dir: Direction } {
	return sort.startsWith("-")
		? { field: sort.slice(1) as SortField, dir: "desc" }
		: { field: sort as SortField, dir: "asc" };
}

export async function listItems(db: Database, input: ListItemsParams): Promise<Page<ItemSummary>> {
	const p = listItemsParamsSchema.parse(input);
	const { field, dir } = parseSort(p.sort);
	const { expr, cast } = sortExpression(field, dir);

	const cursor = p.cursor ? decodeCursor(p.cursor, sortCursorSchema) : null;
	if (p.cursor && (!cursor || cursor.s !== p.sort)) throw new InvalidCursorError();
	const backwards = cursor?.d === "prev";
	// Walking backwards flips both the comparison and the order; rows are reversed again below.
	const effectiveDir: Direction = backwards ? (dir === "asc" ? "desc" : "asc") : dir;

	const where: (SQL | undefined)[] = [eq(items.type, p.type)];
	if (p.genre) where.push(sql`${items.meta} @> ${JSON.stringify({ genres: [p.genre] })}::jsonb`);
	if (p.platform) where.push(sql`${items.meta} @> ${JSON.stringify({ platforms: [p.platform] })}::jsonb`);
	if (p.year) {
		where.push(gte(items.releaseDate, `${p.year}-01-01`), lte(items.releaseDate, `${p.year}-12-31`));
	}
	if (p.minScore !== undefined) where.push(visible, gte(items.score, p.minScore));
	if (p.maxScore !== undefined) where.push(visible, lte(items.score, p.maxScore));
	for (const [key, min] of Object.entries(p.minCategory ?? {})) {
		const column = scoreColumns[Number(key.slice(1)) - 1];
		if (column && min !== undefined) where.push(visible, gte(column, min));
	}
	if (cursor) {
		const op = sql.raw(effectiveDir === "asc" ? ">" : "<");
		where.push(sql`(${expr}, ${items.id}) ${op} (cast(${cursor.v} as ${sql.raw(cast)}), ${cursor.id}::uuid)`);
	}

	const order = effectiveDir === "asc" ? asc : desc;
	const join = withTranslations(p.locale);
	const rows = await db
		.select({ ...summaryColumns, sortValue: sql<string>`(${expr})::text` })
		.from(items)
		.leftJoin(tl, join.tl)
		.leftJoin(ten, join.ten)
		.where(and(...where))
		.orderBy(order(expr), order(items.id))
		.limit(p.limit + 1);

	const hasMore = rows.length > p.limit;
	const pageRows = rows.slice(0, p.limit);
	if (backwards) pageRows.reverse();

	const first = pageRows[0];
	const last = pageRows.at(-1);
	const make = (row: { sortValue: string; id: string } | undefined, d: "next" | "prev") =>
		row ? encodeCursor({ s: p.sort, v: row.sortValue, id: row.id, d }) : null;

	return {
		data: pageRows.map(toSummary),
		nextCursor: backwards || hasMore ? make(last, "next") : null,
		prevCursor: backwards ? (hasMore ? make(first, "prev") : null) : cursor ? make(first, "prev") : null,
	};
}

// ---------------------------------------------------------------------------------------------------------------
// Item detail

async function getItem(db: Database, where: SQL, locale: Locale): Promise<ItemDetail | null> {
	const join = withTranslations(locale);
	const [row] = await db
		.select({
			...summaryColumns,
			externalIds: items.externalIds,
			wikidataQid: items.wikidataQid,
			synopsisLocale: tl.synopsis,
			licenseLocale: tl.synopsisLicense,
			sourceLocale: tl.synopsisSourceUrl,
			synopsisEn: ten.synopsis,
			licenseEn: ten.synopsisLicense,
			sourceEn: ten.synopsisSourceUrl,
		})
		.from(items)
		.leftJoin(tl, join.tl)
		.leftJoin(ten, join.ten)
		.where(where)
		.limit(1);
	if (!row) return null;

	const timeToBeat = row.type === "game" ? await getTimeToBeat(db, row.id) : null;
	const ownSynopsis = row.synopsisLocale !== null;
	const synopsisFallback = !ownSynopsis && row.synopsisEn !== null;
	return {
		...toSummary(row),
		locale,
		synopsis: ownSynopsis ? row.synopsisLocale : row.synopsisEn,
		synopsisLocale: ownSynopsis ? locale : row.synopsisEn !== null ? "en" : null,
		synopsisLicense: ownSynopsis ? row.licenseLocale : row.licenseEn,
		synopsisSourceUrl: httpsUrlOrNull(ownSynopsis ? row.sourceLocale : row.sourceEn),
		translationFallback: locale !== "en" && (row.titleLocale === null || synopsisFallback),
		externalIds: parseExternalIds(row.externalIds, row.id),
		wikidataQid: row.wikidataQid,
		meta: parseMeta(row.meta, row.id),
		timeToBeat,
	};
}

export function getItemBySlug(db: Database, type: ItemType, slug: string, locale: Locale): Promise<ItemDetail | null> {
	return getItem(db, and(eq(items.type, type), eq(items.slug, slug)) as SQL, locale);
}

export function getItemById(db: Database, id: string, locale: Locale): Promise<ItemDetail | null> {
	if (!z.uuid().safeParse(id).success) return Promise.resolve(null);
	return getItem(db, eq(items.id, id), locale);
}

export async function getItemIdBySlug(db: Database, type: ItemType, slug: string): Promise<string | null> {
	const [row] = await db
		.select({ id: items.id })
		.from(items)
		.where(and(eq(items.type, type), eq(items.slug, slug)))
		.limit(1);
	return row?.id ?? null;
}

export type SitemapItem = { type: ItemType; slug: string; updatedAt: string };

export async function countItems(db: Database): Promise<number> {
	const [row] = await db.select({ n: sql<number>`count(*)::int` }).from(items);
	return row?.n ?? 0;
}

/** Stable page of items for the sitemap, ordered by id (uuid v7, so new items land in the last chunk). */
export async function listSitemapItems(db: Database, offset: number, limit: number): Promise<SitemapItem[]> {
	const rows = await db
		.select({ type: items.type, slug: items.slug, updatedAt: items.updatedAt })
		.from(items)
		.orderBy(asc(items.id))
		.offset(offset)
		.limit(limit);
	return rows.map((r) => ({ type: r.type, slug: r.slug, updatedAt: r.updatedAt.toISOString() }));
}

export const LOOKUP_SOURCES = ["imdb", "steam", "igdb", "tmdb", "isbn13", "openlibrary", "wikidata"] as const;
export type LookupSource = (typeof LOOKUP_SOURCES)[number];

/** Exact match on an external id (GIN containment on `external_ids`, unique index for Wikidata). */
export async function getItemIdByExternalId(db: Database, source: LookupSource, id: string): Promise<string | null> {
	const where =
		source === "wikidata"
			? eq(items.wikidataQid, id)
			: sql`${items.externalIds} @> ${JSON.stringify({ [source]: id })}::jsonb`;
	const [row] = await db.select({ id: items.id }).from(items).where(where).limit(1);
	return row?.id ?? null;
}

// ---------------------------------------------------------------------------------------------------------------
// Search — pg_trgm word similarity on titles (typo tolerant) OR full-text match on title + synopsis

// "zleda" vs "Zelda" scores 0.2; the pg_trgm default (0.6) would only catch near-exact words.
const WORD_SIMILARITY_THRESHOLD = "0.2";

export async function searchItems(
	db: Database,
	q: string,
	options: { type?: ItemType; locale: Locale; limit?: number },
): Promise<ItemSummary[]> {
	// Control characters (NUL makes Postgres throw) are never meaningful in a title search: page, autocomplete and
	// /v1/search all reach this function.
	const query = q
		.replace(/\p{Cc}/gu, " ")
		.trim()
		.slice(0, 100);
	if (query.length < 2) return [];
	const limit = Math.min(Math.max(options.limit ?? 20, 1), 50);

	return db.transaction(async (tx) => {
		await tx.execute(
			sql`select set_config('pg_trgm.word_similarity_threshold', ${WORD_SIMILARITY_THRESHOLD}, true)`,
		);
		const tsq = sql`websearch_to_tsquery('simple', ${query})`;
		const ranked = tx
			.select({
				itemId: itemTranslations.itemId,
				rank: sql<number>`max(greatest(word_similarity(${query}, ${itemTranslations.title}), ts_rank(${itemTranslations.search}, ${tsq})))`.as(
					"rank",
				),
			})
			.from(itemTranslations)
			.where(or(sql`${query} <% ${itemTranslations.title}`, sql`${itemTranslations.search} @@ ${tsq}`))
			.groupBy(itemTranslations.itemId)
			.as("ranked");

		const join = withTranslations(options.locale);
		const rows = await tx
			.select(summaryColumns)
			.from(ranked)
			.innerJoin(items, eq(items.id, ranked.itemId))
			.leftJoin(tl, join.tl)
			.leftJoin(ten, join.ten)
			.where(options.type ? eq(items.type, options.type) : undefined)
			.orderBy(desc(ranked.rank), desc(items.popularity), asc(items.id))
			.limit(limit);
		return rows.map(toSummary);
	});
}

// ---------------------------------------------------------------------------------------------------------------
// Facets and side data

async function listMetaValues(db: Database, type: ItemType, key: "genres" | "platforms"): Promise<string[]> {
	const value = sql<string>`jsonb_array_elements_text(${items.meta}->${key})`;
	// Ordinal ORDER BY: with DISTINCT, Postgres rejects the re-parameterized expression as "not in select list".
	const rows = await db
		.selectDistinct({ value })
		.from(items)
		.where(and(eq(items.type, type), sql`jsonb_typeof(${items.meta}->${key}) = 'array'`))
		.orderBy(sql`1`);
	return rows.map((r) => r.value).filter((v) => v.trim().length > 0 && v.length <= 100);
}

export function listGenres(db: Database, type: ItemType): Promise<string[]> {
	return listMetaValues(db, type, "genres");
}

/** Genres with how many items of the type carry each one, most common first (ties by name). */
export async function listGenreCounts(db: Database, type: ItemType): Promise<{ value: string; count: number }[]> {
	const value = sql<string>`jsonb_array_elements_text(${items.meta}->'genres')`;
	const rows = await db
		.select({ value, count: sql<number>`count(*)::int` })
		.from(items)
		.where(and(eq(items.type, type), sql`jsonb_typeof(${items.meta}->'genres') = 'array'`))
		.groupBy(sql`1`)
		.orderBy(sql`2 desc`, sql`1`);
	return rows
		.map((r) => ({ value: r.value, count: Number(r.count) }))
		.filter((r) => r.value.trim().length > 0 && r.value.length <= 100);
}

export function listPlatforms(db: Database, type: ItemType): Promise<string[]> {
	return listMetaValues(db, type, "platforms");
}

export async function getExternalScores(db: Database, itemId: string): Promise<ExternalScore[]> {
	return db
		.select({ source: externalScores.source, value: externalScores.value, url: externalScores.url })
		.from(externalScores)
		.where(eq(externalScores.itemId, itemId))
		.orderBy(externalScores.source);
}

export async function getAffiliateLinks(db: Database, itemId: string): Promise<AffiliateLink[]> {
	const rows = await db
		.select({ store: affiliateLinks.store, url: affiliateLinks.url })
		.from(affiliateLinks)
		.where(and(eq(affiliateLinks.itemId, itemId), eq(affiliateLinks.active, true)))
		.orderBy(affiliateLinks.store);
	return rows.flatMap((r) => {
		const url = httpsUrlOrNull(r.url);
		return url ? [{ store: r.store, url }] : [];
	});
}

// ---------------------------------------------------------------------------------------------------------------
// Public ratings of an item (only public, active profiles)

const ratingCursorSchema = z.object({ t: z.string().max(64), u: z.string().max(128) });

export async function getItemRatingsPage(
	db: Database,
	itemId: string,
	options: { cursor?: string; limit?: number } = {},
): Promise<Page<PublicRating>> {
	const limit = Math.min(Math.max(options.limit ?? 20, 1), 50);
	const cursor = options.cursor ? decodeCursor(options.cursor, ratingCursorSchema) : null;
	if (options.cursor && !cursor) throw new InvalidCursorError();

	const rows = await db
		.select({
			nickname: user.username,
			overall: ratings.overall,
			c1: ratings.c1,
			c2: ratings.c2,
			c3: ratings.c3,
			c4: ratings.c4,
			c5: ratings.c5,
			c6: ratings.c6,
			createdAt: ratings.createdAt,
			cursorTime: sql<string>`${ratings.createdAt}::text`,
			userId: ratings.userId,
		})
		.from(ratings)
		.innerJoin(user, eq(user.id, ratings.userId))
		.where(
			and(
				eq(ratings.itemId, itemId),
				eq(user.profilePrivate, false),
				sql`${user.username} is not null`,
				sql`coalesce(${user.banned}, false) = false`,
				isNull(user.deletionScheduledAt),
				cursor
					? sql`(${ratings.createdAt}, ${ratings.userId}) < (${cursor.t}::timestamptz, ${cursor.u})`
					: undefined,
			),
		)
		.orderBy(desc(ratings.createdAt), desc(ratings.userId))
		.limit(limit + 1);

	const pageRows = rows.slice(0, limit);
	const last = pageRows.at(-1);
	return {
		data: pageRows.map((r) => ({
			nickname: r.nickname ?? "",
			overall: r.overall ?? 0,
			categories: [r.c1, r.c2, r.c3, r.c4, r.c5, r.c6],
			ratedAt: new Date(r.createdAt).toISOString(),
		})),
		nextCursor: rows.length > limit && last ? encodeCursor({ t: last.cursorTime, u: last.userId }) : null,
		prevCursor: null,
	};
}

export async function itemExists(db: Database, itemId: string): Promise<boolean> {
	if (!z.uuid().safeParse(itemId).success) return false;
	const [row] = await db.select({ id: items.id }).from(items).where(eq(items.id, itemId)).limit(1);
	return Boolean(row);
}
