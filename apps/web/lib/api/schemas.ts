import { z } from "zod";
import { getPathname } from "@/i18n/navigation";
import { type Locale, routing } from "@/i18n/routing";
import { LOOKUP_SOURCES } from "@/lib/catalog/repository";
import { type ItemDetail, type ItemSummary, SORTS, type TimeToBeatStat } from "@/lib/catalog/types";
import { CATEGORY_KEYS } from "@/lib/categories";
import { APP_URL } from "@/lib/config";
import { ITEM_TYPES, type ItemType } from "@/lib/db/schema/catalog";
import { itemHref, SECTION_PATHS, type SectionPath } from "@/lib/sections";
import { MIN_REPORTS_FOR_TIME_TO_BEAT } from "@/lib/social/constants";
import { apiDoc } from "./openapi-text";
import { PLAN_IDS } from "./plans";
import { PROBLEM_CODES } from "./problem";

// Single source of truth for the public API: route handlers parse every response through these schemas and
// lib/api/openapi.ts turns the same objects into openapi.json. Wire format is snake_case; dates are ISO 8601.

const score = z.int().min(0).max(100);
const count = z.int().min(0);

export const itemTypeSchema = z.enum(ITEM_TYPES).meta({
	description: apiDoc("schemas.itemType"),
});

const categoryDescription = Object.entries(CATEGORY_KEYS)
	.map(([type, keys]) => `\`${type}\`: ${keys.map((k) => `\`${k}\``).join(", ")}`)
	.join("; ");

export const itemSummarySchema = z
	.object({
		id: z.uuid().meta({
			description: apiDoc("schemas.ItemSummary.id"),
			examples: ["0190a1b2-0000-7000-8000-000000000001"],
		}),
		type: itemTypeSchema,
		slug: z.string().meta({ description: apiDoc("schemas.ItemSummary.slug"), examples: ["the-legend-of-zelda"] }),
		title: z.string().meta({ description: apiDoc("schemas.ItemSummary.title") }),
		release_date: z.iso
			.date()
			.nullable()
			.meta({ description: apiDoc("schemas.ItemSummary.release_date") }),
		year: z.int().nullable(),
		score: score.nullable().meta({
			description: apiDoc("schemas.ItemSummary.score"),
			examples: [87],
		}),
		categories: z.record(z.string(), score.nullable()).meta({
			description: apiDoc("schemas.ItemSummary.categories", { categories: categoryDescription }),
			examples: [{ gameplay: 90, fun: 88, sound: 85, technical: 84, visual: 89, storytelling: 86 }],
		}),
		rating_count: count.meta({ description: apiDoc("schemas.ItemSummary.rating_count") }),
		genres: z.array(z.string()),
		platforms: z.array(z.string()).meta({ description: apiDoc("schemas.ItemSummary.platforms") }),
		cover_url: z.url().nullable(),
		url: z.url().meta({ description: apiDoc("schemas.ItemSummary.url") }),
	})
	.meta({ description: apiDoc("schemas.ItemSummary._") });

export const synopsisSchema = z
	.object({
		text: z.string(),
		locale: z.enum(routing.locales).meta({ description: apiDoc("schemas.Synopsis.locale") }),
		license: z
			.string()
			.nullable()
			.meta({
				description: apiDoc("schemas.Synopsis.license"),
			}),
		source_url: z
			.url()
			.nullable()
			.meta({ description: apiDoc("schemas.Synopsis.source_url") }),
	})
	.meta({ description: apiDoc("schemas.Synopsis._") });

const externalIdsSchema = z
	.object({
		wikidata: z.string(),
		imdb: z.string().optional(),
		igdb: z.string().optional(),
		steam: z.string().optional(),
		tmdb: z.string().optional(),
		isbn13: z.string().optional(),
		openlibrary: z.string().optional(),
		hltb: z.string().optional(),
		goodreads: z.string().optional(),
		metacritic: z.string().optional(),
		rottentomatoes: z.string().optional(),
		epic: z.string().optional(),
		gog: z.string().optional(),
		skoob: z.string().optional(),
	})
	.meta({
		description: apiDoc("schemas.ExternalIds._"),
		examples: [{ wikidata: "Q12345", igdb: "the-legend-of-zelda", steam: "123456" }],
	});

const timeToBeatStatSchema = z.object({
	median_hours: z
		.number()
		.min(0)
		.nullable()
		.meta({
			description: apiDoc("schemas.TimeToBeatStat.median_hours", { minReports: MIN_REPORTS_FOR_TIME_TO_BEAT }),
			examples: [42.5],
		}),
	reports: count.meta({ description: apiDoc("schemas.TimeToBeatStat.reports") }),
});

export const timeToBeatSchema = z
	.object({ main: timeToBeatStatSchema, extras: timeToBeatStatSchema, completionist: timeToBeatStatSchema })
	.meta({
		description: apiDoc("schemas.TimeToBeat._"),
		examples: [
			{
				main: { median_hours: 42.5, reports: 12 },
				extras: { median_hours: 61, reports: 7 },
				completionist: { median_hours: null, reports: 2 },
			},
		],
	});

export const itemDetailSchema = itemSummarySchema
	.extend({
		locale: z.enum(routing.locales),
		translation_fallback: z.boolean().meta({
			description: apiDoc("schemas.ItemDetail.translation_fallback"),
		}),
		synopsis: synopsisSchema.nullable(),
		external_ids: externalIdsSchema,
		trailers: z.array(z.object({ youtube_id: z.string(), title: z.string().nullable() })).meta({
			description: apiDoc("schemas.ItemDetail.trailers"),
		}),
		creators: z.array(z.string()).meta({ description: apiDoc("schemas.ItemDetail.creators") }),
		countries: z.array(z.string()),
		runtime_minutes: z.int().nullable(),
		pages: z.int().nullable(),
		seasons: z.int().nullable(),
		episodes: z.int().nullable(),
		time_to_beat: timeToBeatSchema.nullable().meta({ description: apiDoc("schemas.ItemDetail.time_to_beat") }),
	})
	.meta({ description: apiDoc("schemas.ItemDetail._") });

export const itemListSchema = z
	.object({
		data: z.array(itemSummarySchema),
		next_cursor: z
			.string()
			.nullable()
			.meta({
				description: apiDoc("schemas.ItemList.next_cursor"),
			}),
	})
	.meta({ description: apiDoc("schemas.ItemList._") });

export const searchResultSchema = z
	.object({ data: z.array(itemSummarySchema) })
	.meta({ description: apiDoc("schemas.SearchResult._") });

export const genreListSchema = z
	.object({
		data: z.array(
			z.object({
				slug: z.string().meta({ description: apiDoc("schemas.GenreList.slug"), examples: ["Adventure"] }),
				count: count.meta({ description: apiDoc("schemas.GenreList.count") }),
			}),
		),
	})
	.meta({ description: apiDoc("schemas.GenreList._") });

export const usageSchema = z
	.object({
		plan: z.enum(PLAN_IDS),
		quota: z.object({
			limit: count,
			used: count,
			remaining: count,
			period: z.enum(["lifetime", "month"]).meta({
				description: apiDoc("schemas.Usage.period"),
			}),
			resets_at: z.iso
				.datetime()
				.nullable()
				.meta({ description: apiDoc("schemas.Usage.resets_at") }),
		}),
		credits: z.object({
			balance: count.meta({ description: apiDoc("schemas.Usage.balance") }),
		}),
		rate_limit: z.object({ requests_per_second: count }),
	})
	.meta({ description: apiDoc("schemas.Usage._") });

export const problemSchema = z
	.object({
		type: z.url().meta({ description: apiDoc("schemas.Problem.type") }),
		title: z.string(),
		status: z.int(),
		detail: z.string().meta({ description: apiDoc("schemas.Problem.detail") }),
		code: z.enum(PROBLEM_CODES).meta({ description: apiDoc("schemas.Problem.code") }),
		request_id: z.string().meta({ description: apiDoc("schemas.Problem.request_id") }),
		errors: z
			.array(z.object({ path: z.string(), message: z.string() }))
			.optional()
			.meta({ description: apiDoc("schemas.Problem.errors") }),
		plan: z.enum(PLAN_IDS).optional(),
		limit: z.int().optional(),
	})
	.meta({ description: apiDoc("schemas.Problem._") });

// ---------------------------------------------------------------------------------------------------------------
// Path and query parameters (query values arrive as strings, hence z.coerce)

export const SECTION_TO_TYPE = Object.fromEntries(
	Object.entries(SECTION_PATHS).map(([type, section]) => [section, type]),
) as Record<SectionPath, ItemType>;
export const SECTIONS = Object.values(SECTION_PATHS) as [SectionPath, ...SectionPath[]];

export const typeParamsSchema = z
	.object({ type: z.enum(SECTIONS) })
	.transform((p) => ({ type: SECTION_TO_TYPE[p.type] }));
export const itemParamsSchema = z
	.object({ type: z.enum(SECTIONS), id: z.uuid() })
	.transform((p) => ({ type: SECTION_TO_TYPE[p.type], id: p.id }));

const scoreParam = z.coerce.number().int().min(0).max(100);
const localeParam = z
	.enum(routing.locales)
	.default("en")
	.meta({ description: apiDoc("query.locale") });

function limitParam(maxPageSize: number, upTo = maxPageSize) {
	const max = Math.min(maxPageSize, upTo);
	return z.coerce
		.number()
		.int()
		.min(1)
		.max(max, `Your plan allows at most ${max} results per page.`)
		.default(Math.min(10, max))
		.meta({ description: apiDoc("query.limit", { upTo }) });
}

export function listQuerySchema(type: ItemType, maxPageSize: number) {
	const categoryFilters = Object.fromEntries(
		CATEGORY_KEYS[type].map((key) => [
			`min_${key}`,
			scoreParam.optional().meta({ description: apiDoc("query.min_category", { key }) }),
		]),
	);
	return z.strictObject({
		genre: z
			.string()
			.trim()
			.min(1)
			.max(100)
			.optional()
			.meta({
				description: apiDoc("query.genre"),
			}),
		...(type === "game"
			? {
					platform: z
						.string()
						.trim()
						.min(1)
						.max(100)
						.optional()
						.meta({ description: apiDoc("query.platform") }),
				}
			: {}),
		year: z.coerce
			.number()
			.int()
			.min(1800)
			.max(2100)
			.optional()
			.meta({ description: apiDoc("query.year") }),
		min_score: scoreParam.optional().meta({ description: apiDoc("query.min_score") }),
		max_score: scoreParam.optional().meta({ description: apiDoc("query.max_score") }),
		...categoryFilters,
		sort: z
			.enum(SORTS)
			.default("-popularity")
			.meta({
				description: apiDoc("query.sort"),
				// Readable enum member names in generated SDKs (`SCORE_DESC` instead of `VALUE_2`).
				"x-enum-varnames": SORTS.map((s) =>
					s.startsWith("-") ? `${s.slice(1).toUpperCase()}_DESC` : s.toUpperCase(),
				),
			}),
		limit: limitParam(maxPageSize),
		cursor: z
			.string()
			.max(512)
			.optional()
			.meta({ description: apiDoc("query.cursor") }),
		locale: localeParam,
	});
}

export const detailQuerySchema = z.strictObject({ locale: localeParam });
export const emptyQuerySchema = z.strictObject({});

export const SEARCH_MAX = 50;
export function searchQuerySchema(maxPageSize: number) {
	return z.strictObject({
		q: z
			.string()
			.trim()
			.min(2)
			.max(100)
			.meta({ description: apiDoc("query.q"), examples: ["zelda"] }),
		type: itemTypeSchema.optional().meta({ description: apiDoc("query.type") }),
		limit: limitParam(maxPageSize, SEARCH_MAX),
		locale: localeParam,
	});
}

export const lookupQuerySchema = z.strictObject({
	source: z.enum(LOOKUP_SOURCES).meta({ description: apiDoc("query.source") }),
	id: z
		.string()
		.trim()
		.min(1)
		.max(100)
		.meta({ description: apiDoc("query.id"), examples: ["tt0111161"] }),
	locale: localeParam,
});

// ---------------------------------------------------------------------------------------------------------------
// Repository DTO → wire format

function itemUrl(type: ItemType, slug: string, locale: Locale): string {
	return `${APP_URL}${getPathname({ href: itemHref(type, slug), locale })}`;
}

function categoriesOf(type: ItemType, values: (number | null)[]): Record<string, number | null> {
	return Object.fromEntries(CATEGORY_KEYS[type].map((key, i) => [key, values[i] ?? null]));
}

export function toItemSummary(item: ItemSummary, locale: Locale): z.input<typeof itemSummarySchema> {
	return {
		id: item.id,
		type: item.type,
		slug: item.slug,
		title: item.title,
		release_date: item.releaseDate,
		year: item.year,
		score: item.score,
		categories: categoriesOf(item.type, item.categories),
		rating_count: item.ratingCount,
		genres: item.genres,
		platforms: item.platforms,
		cover_url: item.coverUrl,
		url: itemUrl(item.type, item.slug, locale),
	};
}

const toTimeToBeatStat = (s: TimeToBeatStat): z.input<typeof timeToBeatStatSchema> => ({
	median_hours: s.medianHours,
	reports: s.reports,
});

export function toItemDetail(item: ItemDetail, locale: Locale): z.input<typeof itemDetailSchema> {
	const synopsisLocale = routing.locales.find((l) => l === item.synopsisLocale);
	return {
		...toItemSummary(item, locale),
		locale,
		translation_fallback: item.translationFallback,
		synopsis:
			item.synopsis && synopsisLocale
				? {
						text: item.synopsis,
						locale: synopsisLocale,
						license: item.synopsisLicense,
						source_url: item.synopsisSourceUrl,
					}
				: null,
		external_ids: { wikidata: item.wikidataQid, ...item.externalIds },
		trailers: item.meta.trailers.map((t) => ({ youtube_id: t.youtubeId, title: t.title ?? null })),
		creators: item.meta.creators ?? [],
		countries: item.meta.countries ?? [],
		runtime_minutes: item.meta.runtimeMinutes ?? null,
		pages: item.meta.pages ?? null,
		seasons: item.meta.seasons ?? null,
		episodes: item.meta.episodes ?? null,
		time_to_beat: item.timeToBeat
			? {
					main: toTimeToBeatStat(item.timeToBeat.main),
					extras: toTimeToBeatStat(item.timeToBeat.extras),
					completionist: toTimeToBeatStat(item.timeToBeat.completionist),
				}
			: null,
	};
}
