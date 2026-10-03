import { sql } from "drizzle-orm";
import {
	boolean,
	customType,
	date,
	index,
	integer,
	jsonb,
	pgEnum,
	pgTable,
	primaryKey,
	real,
	smallint,
	text,
	timestamp,
	uniqueIndex,
	uuid,
} from "drizzle-orm/pg-core";

export const itemType = pgEnum("item_type", ["game", "movie", "series", "book"]);
export const ITEM_TYPES = itemType.enumValues;
export type ItemType = (typeof ITEM_TYPES)[number];

export const locale = pgEnum("locale", ["en", "pt-BR", "es"]);

const tsvector = customType<{ data: string }>({ dataType: () => "tsvector" });

export type ExternalIds = Partial<
	Record<
		| "imdb"
		| "igdb"
		| "steam"
		| "tmdb"
		| "isbn13"
		| "openlibrary"
		| "hltb"
		| "goodreads"
		| "metacritic"
		| "rottentomatoes"
		| "epic"
		| "gog"
		| "skoob",
		string
	>
>;

export type Trailer = { youtubeId: string; title?: string };

export type ItemMeta = {
	genres?: string[];
	platforms?: string[];
	creators?: string[];
	runtimeMinutes?: number;
	pages?: number;
	seasons?: number;
	episodes?: number;
	countries?: string[];
	trailers?: Trailer[];
	coverUrl?: string;
};

export const items = pgTable(
	"items",
	{
		id: uuid()
			.primaryKey()
			.$defaultFn(() => Bun.randomUUIDv7()),
		type: itemType().notNull(),
		slug: text().notNull(),
		wikidataQid: text().notNull().unique(),
		releaseDate: date(),
		externalIds: jsonb().$type<ExternalIds>().notNull().default({}),
		meta: jsonb().$type<ItemMeta>().notNull().default({}),
		popularity: real().notNull().default(0),
		ratingCount: integer().notNull().default(0),
		score: smallint(),
		scoreC1: smallint(),
		scoreC2: smallint(),
		scoreC3: smallint(),
		scoreC4: smallint(),
		scoreC5: smallint(),
		scoreC6: smallint(),
		createdAt: timestamp({ withTimezone: true }).notNull().defaultNow(),
		updatedAt: timestamp({ withTimezone: true })
			.notNull()
			.defaultNow()
			.$onUpdate(() => new Date()),
	},
	(t) => [
		uniqueIndex("items_type_slug_idx").on(t.type, t.slug),
		index("items_type_score_idx").on(t.type, t.score),
		index("items_type_popularity_idx").on(t.type, t.popularity),
		index("items_type_release_idx").on(t.type, t.releaseDate),
		index("items_meta_gin_idx").using("gin", t.meta),
		index("items_external_ids_gin_idx").using("gin", t.externalIds),
	],
);

export const itemTranslations = pgTable(
	"item_translations",
	{
		itemId: uuid()
			.notNull()
			.references(() => items.id, { onDelete: "cascade" }),
		locale: locale().notNull(),
		title: text().notNull(),
		synopsis: text(),
		synopsisLicense: text(),
		synopsisSourceUrl: text(),
		search: tsvector().generatedAlwaysAs(
			sql`to_tsvector('simple', coalesce(title, '') || ' ' || coalesce(synopsis, ''))`,
		),
	},
	(t) => [
		primaryKey({ columns: [t.itemId, t.locale] }),
		index("item_translations_search_idx").using("gin", t.search),
		index("item_translations_title_trgm_idx").using("gin", sql`${t.title} gin_trgm_ops`),
	],
);

export const externalScoreSource = pgEnum("external_score_source", [
	"metacritic",
	"metacritic_user",
	"rottentomatoes",
	"rottentomatoes_audience",
	"imdb",
	"steam",
	"goodreads",
	"hltb_main",
	"hltb_complete",
]);

export const externalScores = pgTable(
	"external_scores",
	{
		itemId: uuid()
			.notNull()
			.references(() => items.id, { onDelete: "cascade" }),
		source: externalScoreSource().notNull(),
		value: text().notNull(),
		url: text(),
		fetchedAt: timestamp({ withTimezone: true }).notNull().defaultNow(),
	},
	(t) => [primaryKey({ columns: [t.itemId, t.source] })],
);

export const affiliateStore = pgEnum("affiliate_store", ["amazon", "mercadolivre", "steam", "gog"]);

export const affiliateLinks = pgTable(
	"affiliate_links",
	{
		itemId: uuid()
			.notNull()
			.references(() => items.id, { onDelete: "cascade" }),
		store: affiliateStore().notNull(),
		url: text().notNull(),
		active: boolean().notNull().default(true),
	},
	(t) => [primaryKey({ columns: [t.itemId, t.store] })],
);

export const sourceRecords = pgTable(
	"source_records",
	{
		source: text().notNull(),
		externalId: text().notNull(),
		payload: jsonb().notNull(),
		hash: text().notNull(),
		fetchedAt: timestamp({ withTimezone: true }).notNull().defaultNow(),
		normalizedAt: timestamp({ withTimezone: true }),
		itemId: uuid().references(() => items.id, { onDelete: "set null" }),
	},
	(t) => [
		primaryKey({ columns: [t.source, t.externalId] }),
		index("source_records_pending_idx")
			.on(t.source)
			.where(sql`normalized_at is null or normalized_at < fetched_at`),
	],
);
