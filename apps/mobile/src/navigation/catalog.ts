export const CATALOG_TYPES = ["games", "movies", "series", "books"] as const;

export type CatalogType = (typeof CATALOG_TYPES)[number];

export function isCatalogType(value: string): value is CatalogType {
	return (CATALOG_TYPES as readonly string[]).includes(value);
}

export function firstParam(value: string | string[] | undefined): string {
	if (Array.isArray(value)) return value[0] ?? "";
	return value ?? "";
}

// API items carry the plural route segment as `type`; catalogue keys are singular.
export const SINGULAR: Record<CatalogType, string> = {
	games: "game",
	movies: "movie",
	series: "series",
	books: "book",
};
