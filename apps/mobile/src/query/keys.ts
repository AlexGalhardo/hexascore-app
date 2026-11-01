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

export const queryKeys = {
	trending: (type: CatalogType, locale: AppLocale) => [TRENDING_METHOD[type], { locale }] as const,
};
