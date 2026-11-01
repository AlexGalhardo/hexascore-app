import type { AppLocale } from "./i18n/locale";

/** SDD 10 §7.6: the paths `android.intentFilters` verifies on hexascore.app. */
export const APP_LINK_PREFIXES = [
	"/games",
	"/movies",
	"/series",
	"/books",
	"/entities",
	"/lists",
	"/pt-br",
	"/es",
] as const;

const LOCALE_SEGMENTS: Record<string, AppLocale> = { "pt-br": "pt-BR", es: "es" };
const SCREENS = new Set(["games", "movies", "series", "books", "entities", "lists"]);

export type DeepLink = { path: string; locale: AppLocale | null };

/**
 * Maps a web URL to an app route: `/{,pt-br/,es/}{type}/{slug}`, `/entities/{id}` and `/lists/{id}`
 * keep their path (the route segments are the web's, FR-10-7); the locale prefix is returned
 * apart so the caller applies it for the session only. Anything else opens home.
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
	return { path: `/${segments.join("/")}${search}`, locale };
}
