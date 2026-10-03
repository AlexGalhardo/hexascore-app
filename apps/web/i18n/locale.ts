import { notFound } from "next/navigation";
import { hasLocale } from "next-intl";
import { LOCALE_PREFIXES, type Locale, routing } from "./routing";

/**
 * Narrows the `[locale]` param or 404s. Pages must not cast it: the layout's `hasLocale` guard renders in parallel
 * with the page, so `/favicon.ico` (→ locale "favicon.ico") used to reach catalog queries and throw a ZodError.
 */
export function asLocale(value: string): Locale {
	if (!hasLocale(routing.locales, value)) notFound();
	return value;
}

const PREFIXES = Object.entries(LOCALE_PREFIXES) as [keyof typeof LOCALE_PREFIXES, string][];

/**
 * Locale of a request outside a `[locale]` route (e.g. /api/auth/*): the next-intl cookie, then the locale prefix
 * of the page that made the call (same-origin Referer keeps the path), else the default locale.
 */
export function localeFromRequest(request?: Request): Locale {
	const raw = request?.headers.get("cookie")?.match(/(?:^|;\s*)locale=([^;]+)/)?.[1];
	const cookie = raw ? decodeURIComponent(raw) : undefined;
	if (hasLocale(routing.locales, cookie)) return cookie;
	const referer = request?.headers.get("referer");
	const path = referer && URL.canParse(referer) ? new URL(referer).pathname.toLowerCase() : "";
	const match = PREFIXES.find(([, prefix]) => path === prefix || path.startsWith(`${prefix}/`));
	return match?.[0] ?? routing.defaultLocale;
}
