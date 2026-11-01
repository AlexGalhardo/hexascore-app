import type { ReadClient } from "../api/client";
import type { AppLocale } from "../i18n/locale";
import type { CatalogType } from "../navigation/catalog";

// Element 0 of every key is the ReadClient method: the persister allowlist matches on it.
export const TRENDING_METHOD = {
	games: "listGameTrending",
	movies: "listMovieTrending",
	series: "listSeriesTrending",
	books: "listBookTrending",
} as const satisfies Record<CatalogType, keyof ReadClient>;

export const GENRE_METHOD = {
	games: "listGameGenres",
	movies: "listMovieGenres",
	series: "listSeriesGenres",
	books: "listBookGenres",
} as const satisfies Record<CatalogType, keyof ReadClient>;

export type ListingFilter = { sort: string; genre?: string };

export const queryKeys = {
	trending: (type: CatalogType, locale: AppLocale) => [TRENDING_METHOD[type], { locale }] as const,
	genres: (type: CatalogType) => [GENRE_METHOD[type]] as const,
	// The cursor is the page param, not part of the key: a new sort or genre is a new key, so paging restarts.
	listing: (type: CatalogType, filter: ListingFilter, locale: AppLocale) =>
		["listItems", { type, ...filter, locale }] as const,
	item: (type: CatalogType, id: string, locale: AppLocale) =>
		["getItem", { type, id, locale }] as const,
	editions: (id: string) => ["listEditions", { id }] as const,
	itemMe: (id: string) => ["me", { id }] as const,
};
