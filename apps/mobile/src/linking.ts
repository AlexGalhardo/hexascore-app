import type { AppLocale } from "./i18n/locale";

const SCREEN_LIST = ["games", "movies", "series", "books", "entities", "lists"] as const;
const LOCALE_PREFIXES = ["", "/pt-br", "/es"] as const;

/**
 * SDD 10 §7.6: the paths `android.intentFilters` verifies on hexascore.app, as Android
 * `pathAdvancedPattern`s (API 31+; older devices open the browser). `pathPrefix` would also claim the
 * entity pages under a category (`/games/platforms/playstation-5`) that the app has no screen for,
 * so only the single-segment pages are listed. app.config.ts inlines the same list (Expo loads it
 * without TS module resolution); linking.test.ts keeps both equal.
 */
export const APP_LINK_PATTERNS: readonly string[] = [
	...LOCALE_PREFIXES.flatMap((prefix) =>
		SCREEN_LIST.map((screen) => `${prefix}/${screen}/[^/]+/?`),
	),
	"/pt-br/?",
	"/es/?",
];

const LOCALE_SEGMENTS: Record<string, AppLocale> = { "pt-br": "pt-BR", es: "es" };
const SCREENS = new Set<string>(SCREEN_LIST);

export type DeepLink = { path: string; locale: AppLocale | null };

/**
 * Maps a web URL to an app route: `/{,pt-br/,es/}{type}/{slug}`, `/entities/{id}` and `/lists/{id}`
 * keep their path (the route segments are the web's, FR-10-7); the locale prefix is returned
 * apart so the caller applies it for the session only. Anything else, including the entity pages under a
 * category (`/games/platforms/playstation-5`), opens home.
 */
export function resolveDeepLink(url: string): DeepLink {
	// No `URL`: React Native's implementation leaves pathname unimplemented.
	const match = /^(?:([a-z][a-z0-9+.-]*):\/\/([^/?#]*))?([^?#]*)(\?[^#]*)?/i.exec(url);
	const [, scheme = "", host = "", rest = "", search = ""] = match ?? [];
	// `hexascore://series/dark` carries the first segment in the host position.
	const custom = scheme !== "" && !/^https?$/i.test(scheme);
	const path = custom ? `/${host}${rest}` : rest;

	const segments = path.split("/").filter(Boolean);
	const locale = LOCALE_SEGMENTS[segments[0]?.toLowerCase() ?? ""] ?? null;
	if (locale) segments.shift();
	if (segments.length === 0 || !SCREENS.has(segments[0] ?? "")) return { path: "/", locale };
	// Deeper web pages under a section (`/games/platforms/playstation-5`, changes, editions) have no screen.
	if (segments.length > 2) return { path: "/", locale };
	return { path: `/${segments.join("/")}${search}`, locale };
}
