import { z } from "zod";
import type { ExternalIds, ItemMeta, Trailer } from "@/lib/db/schema/catalog";

const text = z.string().trim().min(1).max(200);
const textList = z.array(z.unknown()).transform((list) => list.filter((v) => text.safeParse(v).success) as string[]);
const positiveInt = z.number().int().positive().max(1_000_000);

export const trailerSchema = z.object({
	youtubeId: z.string().regex(/^[A-Za-z0-9_-]{11}$/),
	title: text.optional(),
});

const metaFields = {
	genres: textList,
	platforms: textList,
	creators: textList,
	countries: textList,
	runtimeMinutes: positiveInt,
	pages: positiveInt,
	seasons: positiveInt,
	episodes: positiveInt,
	trailers: z.array(z.unknown()).transform((list) =>
		list.flatMap((v) => {
			const r = trailerSchema.safeParse(v);
			return r.success ? [r.data] : [];
		}),
	),
	coverUrl: z.url({ protocol: /^https$/ }),
	coverThumbUrl: z.url({ protocol: /^https$/ }),
	coverAttribution: z.object({
		source: z.string().max(200),
		license: z.string().max(200).optional(),
		author: z.string().max(300).optional(),
		url: z.url({ protocol: /^https$/ }).optional(),
	}),
} satisfies Record<keyof ItemMeta, z.ZodType>;

// Ids are interpolated into outbound URLs, so each one is constrained to the provider's real format.
const externalIdFields = {
	imdb: z.string().regex(/^tt\d{7,10}$/),
	igdb: z.string().regex(/^[a-z0-9-]{1,100}$/),
	steam: z.string().regex(/^\d{1,10}$/),
	tmdb: z.string().regex(/^(movie|tv)\/\d{1,10}$|^\d{1,10}$/),
	isbn13: z.string().regex(/^97[89]\d{10}$/),
	openlibrary: z.string().regex(/^OL\d{1,10}[WM]$/),
	hltb: z.string().regex(/^\d{1,10}$/),
	goodreads: z.string().regex(/^\d{1,12}$/),
	metacritic: z.string().regex(/^[a-z0-9-]{1,150}$/),
	rottentomatoes: z.string().regex(/^[a-z0-9_-]{1,150}$/),
	epic: z.string().regex(/^[a-z0-9-]{1,150}$/),
	gog: z.string().regex(/^[a-z0-9_-]{1,150}$/),
	skoob: z.string().regex(/^\d{1,12}$/),
} satisfies Record<keyof ExternalIds, z.ZodType>;

// Lenient: a bad field is dropped (and logged) instead of failing the whole page; jsonb is written by the
// Python ingestion, so the web app treats it as untrusted input.
function parseLenient<T extends Record<string, unknown>>(
	fields: Record<string, z.ZodType>,
	raw: unknown,
	context: string,
): Partial<T> {
	if (typeof raw !== "object" || raw === null || Array.isArray(raw)) {
		if (raw !== null && raw !== undefined) console.warn(`[catalog] ${context}: expected an object, dropped`);
		return {};
	}
	const out: Record<string, unknown> = {};
	const dropped: string[] = [];
	for (const [key, value] of Object.entries(raw)) {
		const schema = fields[key];
		if (!schema || value === null || value === undefined) continue;
		const parsed = schema.safeParse(value);
		if (parsed.success) out[key] = parsed.data;
		else dropped.push(key);
	}
	if (dropped.length) console.warn(`[catalog] ${context}: dropped invalid ${dropped.join(", ")}`);
	return out as Partial<T>;
}

export function parseMeta(raw: unknown, itemId: string): ItemMeta & { trailers: Trailer[] } {
	const meta = parseLenient<ItemMeta>(metaFields, raw, `items.meta ${itemId}`);
	return { ...meta, trailers: meta.trailers ?? [] };
}

export function parseExternalIds(raw: unknown, itemId: string): ExternalIds {
	return parseLenient<ExternalIds>(externalIdFields, raw, `items.external_ids ${itemId}`);
}
