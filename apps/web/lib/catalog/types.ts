import { z } from "zod";
import { routing } from "@/i18n/routing";
import { type ExternalIds, ITEM_TYPES, type ItemMeta, type ItemType, type Trailer } from "@/lib/db/schema/catalog";
import type { TimeToBeatStyle } from "@/lib/social/constants";

// DTOs are plain JSON (dates as ISO strings) so they survive unstable_cache and route-handler serialization.

export const SORTS = [
	"-popularity",
	"popularity",
	"-score",
	"score",
	"-release_date",
	"release_date",
	"-rating_count",
	"rating_count",
] as const;
export type Sort = (typeof SORTS)[number];
export const DEFAULT_SORT: Sort = "-popularity";

export const CATEGORY_COLUMNS = ["c1", "c2", "c3", "c4", "c5", "c6"] as const;
export type CategoryColumn = (typeof CATEGORY_COLUMNS)[number];

const score = z.number().int().min(0).max(100);

export const listItemsParamsSchema = z.object({
	type: z.enum(ITEM_TYPES),
	locale: z.enum(routing.locales),
	genre: z.string().trim().min(1).max(100).optional(),
	platform: z.string().trim().min(1).max(100).optional(),
	year: z.number().int().min(1800).max(2100).optional(),
	minScore: score.optional(),
	maxScore: score.optional(),
	minCategory: z.partialRecord(z.enum(CATEGORY_COLUMNS), score).optional(),
	sort: z.enum(SORTS).default(DEFAULT_SORT),
	limit: z.number().int().min(1).max(100).default(25),
	cursor: z.string().max(512).optional(),
});
export type ListItemsParams = z.input<typeof listItemsParamsSchema>;

export type ItemSummary = {
	id: string;
	type: ItemType;
	slug: string;
	title: string;
	releaseDate: string | null;
	year: number | null;
	/** null when the item has fewer than MIN_RATINGS_FOR_SCORE ratings. */
	score: number | null;
	categories: (number | null)[];
	ratingCount: number;
	genres: string[];
	platforms: string[];
	coverUrl: string | null;
};

export type Page<T> = { data: T[]; nextCursor: string | null; prevCursor: string | null };

export type ItemDetail = ItemSummary & {
	locale: string;
	synopsis: string | null;
	/** Locale the synopsis is written in (the requested one or the English fallback). */
	synopsisLocale: string | null;
	synopsisLicense: string | null;
	synopsisSourceUrl: string | null;
	/** True when the requested locale had no title/synopsis and English was used instead. */
	translationFallback: boolean;
	externalIds: ExternalIds;
	wikidataQid: string;
	meta: ItemMeta & { trailers: Trailer[] };
	/** Games only (null for other types). */
	timeToBeat: TimeToBeat | null;
};

/** `medianHours` is null below MIN_REPORTS_FOR_TIME_TO_BEAT reports for that style. */
export type TimeToBeatStat = { medianHours: number | null; reports: number };
export type TimeToBeat = Record<TimeToBeatStyle, TimeToBeatStat>;

export type ExternalScore = { source: string; value: string; url: string | null };
export type AffiliateLink = { store: string; url: string };

export type PublicRating = {
	nickname: string;
	overall: number;
	categories: number[];
	ratedAt: string;
};
