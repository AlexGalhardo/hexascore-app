import { z } from "zod";
import { getPathname } from "@/i18n/navigation";
import { type Locale, routing } from "@/i18n/routing";
import { LOOKUP_SOURCES } from "@/lib/catalog/repository";
import { type ItemDetail, type ItemSummary, SORTS } from "@/lib/catalog/types";
import { CATEGORY_KEYS } from "@/lib/categories";
import { APP_URL } from "@/lib/config";
import { ITEM_TYPES, type ItemType } from "@/lib/db/schema/catalog";
import { itemHref, SECTION_PATHS, type SectionPath } from "@/lib/sections";
import { PLAN_IDS } from "./plans";

// Single source of truth for the public API: route handlers parse every response through these schemas and
// lib/api/openapi.ts turns the same objects into openapi.json. Wire format is snake_case; dates are ISO 8601.

const score = z.int().min(0).max(100);
const count = z.int().min(0);

export const itemTypeSchema = z.enum(ITEM_TYPES).meta({
	description: "Kind of item. List endpoints use the plural path (`/games`, `/movies`, `/series`, `/books`).",
});

const categoryDescription = Object.entries(CATEGORY_KEYS)
	.map(([type, keys]) => `\`${type}\`: ${keys.map((k) => `\`${k}\``).join(", ")}`)
	.join("; ");

export const itemSummarySchema = z
	.object({
		id: z.uuid().meta({ description: "Stable item id.", examples: ["0190a1b2-0000-7000-8000-000000000001"] }),
		type: itemTypeSchema,
		slug: z.string().meta({ description: "URL-safe name, unique per type.", examples: ["the-legend-of-zelda"] }),
		title: z.string().meta({ description: "Title in the requested locale, English when missing." }),
		release_date: z.iso.date().nullable().meta({ description: "First release (YYYY-MM-DD)." }),
		year: z.int().nullable(),
		score: score.nullable().meta({
			description: "Community score 0–100 (mean of the six categories). `null` until the item has 5 ratings.",
			examples: [87],
		}),
		categories: z.record(z.string(), score.nullable()).meta({
			description: `Average per category, 0–100, \`null\` under 5 ratings. Keys by type — ${categoryDescription}.`,
			examples: [{ gameplay: 90, fun: 88, sound: 85, technical: 84, visual: 89, storytelling: 86 }],
		}),
		rating_count: count.meta({ description: "Number of community ratings." }),
		genres: z.array(z.string()),
		platforms: z.array(z.string()).meta({ description: "Platforms (games only; empty for other types)." }),
		cover_url: z.url().nullable(),
		url: z.url().meta({ description: "Public page of the item on the website, in the requested locale." }),
	})
	.meta({ description: "Compact item used in lists and search results." });

export const synopsisSchema = z
	.object({
		text: z.string(),
		locale: z.enum(routing.locales).meta({ description: "Locale the text is written in." }),
		license: z.string().nullable().meta({
			description: "License of the text (e.g. `CC BY-SA 4.0`). Attribute it when you display the synopsis.",
		}),
		source_url: z.url().nullable().meta({ description: "Where the text comes from (for attribution)." }),
	})
	.meta({ description: "Synopsis with the license you must honour when showing it." });

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
		description: "Identifiers of the item in other databases. Only known ids are present.",
		examples: [{ wikidata: "Q12345", igdb: "the-legend-of-zelda", steam: "123456" }],
	});

export const itemDetailSchema = itemSummarySchema
	.extend({
		locale: z.enum(routing.locales),
		translation_fallback: z.boolean().meta({
			description: "True when the requested locale had no title or synopsis and English was used.",
		}),
		synopsis: synopsisSchema.nullable(),
		external_ids: externalIdsSchema,
		trailers: z.array(z.object({ youtube_id: z.string(), title: z.string().nullable() })).meta({
			description: "YouTube video ids; embed as `https://www.youtube-nocookie.com/embed/{youtube_id}`.",
		}),
		creators: z.array(z.string()).meta({ description: "Developers, directors, showrunners or authors." }),
		countries: z.array(z.string()),
		runtime_minutes: z.int().nullable(),
		pages: z.int().nullable(),
		seasons: z.int().nullable(),
		episodes: z.int().nullable(),
	})
	.meta({ description: "Full item: metadata, scores, external ids, trailers and synopsis." });

export const itemListSchema = z
	.object({
		data: z.array(itemSummarySchema),
		next_cursor: z.string().nullable().meta({
			description: "Pass as `cursor` to get the next page. `null` on the last page.",
		}),
	})
	.meta({ description: "One page of items." });

export const searchResultSchema = z
	.object({ data: z.array(itemSummarySchema) })
	.meta({ description: "Best matches first. Search is not paginated." });

export const genreListSchema = z
	.object({
		data: z.array(
			z.object({
				slug: z.string().meta({ description: "Exact value to pass as `genre`.", examples: ["Adventure"] }),
				count: count.meta({ description: "Items of this type with the genre." }),
			}),
		),
	})
	.meta({ description: "Genres of a type, most common first." });

export const usageSchema = z
	.object({
		plan: z.enum(PLAN_IDS),
		quota: z.object({
			limit: count,
			used: count,
			remaining: count,
			period: z.enum(["lifetime", "month"]).meta({
				description: "`lifetime` on Free (100 requests in total); `month` on paid plans (UTC calendar month).",
			}),
			resets_at: z.iso.datetime().nullable().meta({ description: "Next reset; `null` for lifetime quotas." }),
		}),
		credits: z.object({
			balance: count.meta({ description: "Prepaid requests used after the quota runs out (paid plans)." }),
		}),
		rate_limit: z.object({ requests_per_second: count }),
	})
	.meta({ description: "Quota, credits and rate limit of the key's account. This call counts as one request." });

export const PROBLEM_CODES = [
	"bad_request",
	"unauthorized",
	"payment_required",
	"forbidden",
	"not_found",
	"rate_limited",
	"internal_error",
	"service_unavailable",
] as const;

export const problemSchema = z
	.object({
		type: z.url().meta({ description: "Link to the documentation of this error." }),
		title: z.string(),
		status: z.int(),
		detail: z.string().meta({ description: "What went wrong and how to fix it." }),
		code: z.enum(PROBLEM_CODES).meta({ description: "Stable machine-readable error code." }),
		request_id: z.string().meta({ description: "Quote it when contacting support." }),
		errors: z
			.array(z.object({ path: z.string(), message: z.string() }))
			.optional()
			.meta({ description: "Invalid parameters (only on `bad_request`)." }),
		plan: z.enum(PLAN_IDS).optional(),
		limit: z.int().optional(),
	})
	.meta({ description: "RFC 9457 problem details (`application/problem+json`)." });

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
	.meta({ description: "Language of titles and synopses (`en`, `pt-BR`, `es`). Falls back to English." });

function limitParam(maxPageSize: number, upTo = maxPageSize) {
	const max = Math.min(maxPageSize, upTo);
	return z.coerce
		.number()
		.int()
		.min(1)
		.max(max, `Your plan allows at most ${max} results per page.`)
		.default(Math.min(10, max))
		.meta({ description: `Page size, 1–${upTo}. Free plan: up to 10.` });
}

export function listQuerySchema(type: ItemType, maxPageSize: number) {
	const categoryFilters = Object.fromEntries(
		CATEGORY_KEYS[type].map((key) => [
			`min_${key}`,
			scoreParam.optional().meta({ description: `Minimum \`${key}\` average (0–100).` }),
		]),
	);
	return z.strictObject({
		genre: z.string().trim().min(1).max(100).optional().meta({
			description: "Exact genre, as returned by the genres endpoint.",
		}),
		...(type === "game"
			? { platform: z.string().trim().min(1).max(100).optional().meta({ description: "Exact platform name." }) }
			: {}),
		year: z.coerce.number().int().min(1800).max(2100).optional().meta({ description: "Release year." }),
		min_score: scoreParam.optional().meta({ description: "Minimum community score (hides unscored items)." }),
		max_score: scoreParam.optional().meta({ description: "Maximum community score (hides unscored items)." }),
		...categoryFilters,
		sort: z.enum(SORTS).default("-popularity").meta({
			description: "Sort field; prefix with `-` for descending. Unscored/undated items always come last.",
		}),
		limit: limitParam(maxPageSize),
		cursor: z.string().max(512).optional().meta({ description: "`next_cursor` from the previous page." }),
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
			.meta({ description: "Text to search; typo tolerant.", examples: ["zelda"] }),
		type: itemTypeSchema
			.optional()
			.meta({ description: "Restrict to one type (`game`, `movie`, `series`, `book`)." }),
		limit: limitParam(maxPageSize, SEARCH_MAX),
		locale: localeParam,
	});
}

export const lookupQuerySchema = z.strictObject({
	source: z.enum(LOOKUP_SOURCES).meta({ description: "Database the id belongs to." }),
	id: z
		.string()
		.trim()
		.min(1)
		.max(100)
		.meta({ description: "Id in that database.", examples: ["tt0111161"] }),
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
	};
}
