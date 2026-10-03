import { z } from "zod";
import { CATEGORY_KEYS } from "@/lib/categories";
import type { ItemType } from "@/lib/db/schema/catalog";
import { CATEGORY_COLUMNS, type ListItemsParams, SORTS, type Sort } from "./types";

export type RawSearchParams = Record<string, string | string[] | undefined>;

// URL params are user input from a plain GET form: anything invalid is ignored (the field resets), never a 500.
const optional = <T extends z.ZodType>(schema: T) =>
	z.preprocess((v) => (Array.isArray(v) ? v[0] : v === "" ? undefined : v), schema.optional()).catch(undefined);

const text = optional(z.string().trim().min(1).max(100));
const score = optional(z.coerce.number().int().min(0).max(100));

const baseSchema = z.object({
	genre: text,
	platform: text,
	year: optional(z.coerce.number().int().min(1800).max(2100)),
	min_score: score,
	max_score: score,
	sort: optional(z.enum(SORTS)),
	cursor: optional(
		z
			.string()
			.max(512)
			.regex(/^[A-Za-z0-9_-]+$/),
	),
});

export type CatalogFilters = {
	params: Omit<ListItemsParams, "type" | "locale">;
	/** Normalized values to re-fill the form and build pagination links (cursor excluded). */
	query: Record<string, string>;
};

export function parseCatalogSearchParams(type: ItemType, raw: RawSearchParams): CatalogFilters {
	const base = baseSchema.parse(raw);
	if (type !== "game") base.platform = undefined;
	const query: Record<string, string> = {};
	const minCategory: NonNullable<ListItemsParams["minCategory"]> = {};
	CATEGORY_KEYS[type].forEach((key, i) => {
		const value = score.parse(raw[`min_${key}`]);
		if (value === undefined) return;
		minCategory[CATEGORY_COLUMNS[i] as (typeof CATEGORY_COLUMNS)[number]] = value;
		query[`min_${key}`] = String(value);
	});
	for (const key of ["genre", "platform", "year", "min_score", "max_score", "sort"] as const) {
		if (base[key] !== undefined) query[key] = String(base[key]);
	}
	return {
		params: {
			genre: base.genre,
			platform: base.platform,
			year: base.year,
			minScore: base.min_score,
			maxScore: base.max_score,
			minCategory,
			sort: base.sort as Sort | undefined,
			cursor: base.cursor,
		},
		query,
	};
}

export function sortMessageKey(sort: Sort): string {
	return sort.startsWith("-") ? `${sort.slice(1)}_desc` : `${sort}_asc`;
}
