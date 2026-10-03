import type { MetadataRoute } from "next";
import { LOCALE_PREFIXES, type Locale, routing } from "@/i18n/routing";
import type { SitemapItem } from "@/lib/catalog/repository";
import { itemHref } from "@/lib/sections";

/** Items per sitemap file: Google allows 50k URLs; each entry also carries 3 hreflang alternates. */
export const SITEMAP_CHUNK = 40_000;

export const STATIC_PATHS = [
	"/",
	"/games",
	"/movies",
	"/series",
	"/books",
	"/pricing",
	"/contact",
	"/terms",
	"/policy",
	"/docs",
	"/docs/authentication",
	"/docs/pagination",
	"/docs/rate-limits",
	"/docs/errors",
	"/docs/changelog",
	"/docs/reference",
] as const;

// Private, per-user or thin pages; internal search results are kept out of the index as Google recommends.
const DISALLOWED = [
	"/admin",
	"/profile",
	"/feed",
	"/onboarding",
	"/search",
	"/verify-email",
	"/reset-password",
	"/recover-password",
];

export function localizedUrl(base: string, locale: Locale, path: string): string {
	const prefix = locale === "en" ? "" : LOCALE_PREFIXES[locale];
	return `${base}${prefix}${path === "/" && prefix ? "" : path}`;
}

/** One entry per page (default locale URL) with every locale as an hreflang alternate. */
export function sitemapEntry(base: string, path: string, lastModified?: string): MetadataRoute.Sitemap[number] {
	const languages = Object.fromEntries(routing.locales.map((l) => [l, localizedUrl(base, l, path)]));
	return {
		url: localizedUrl(base, routing.defaultLocale, path),
		...(lastModified ? { lastModified } : {}),
		alternates: { languages: { ...languages, "x-default": languages[routing.defaultLocale] } },
	};
}

export function itemEntries(base: string, rows: SitemapItem[]): MetadataRoute.Sitemap {
	return rows.map((r) => sitemapEntry(base, itemHref(r.type, r.slug), r.updatedAt));
}

/** Sitemap ids: 0 = static pages, 1..n = item chunks. */
export function sitemapIds(itemCount: number): number[] {
	return Array.from({ length: 1 + Math.ceil(itemCount / SITEMAP_CHUNK) }, (_, i) => i);
}

export function disallowedPaths(): string[] {
	const prefixes = ["", ...Object.values(LOCALE_PREFIXES)];
	return ["/api/", ...prefixes.flatMap((p) => DISALLOWED.map((d) => `${p}${d}`))];
}
