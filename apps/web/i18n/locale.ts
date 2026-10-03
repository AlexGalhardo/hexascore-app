import { notFound } from "next/navigation";
import { hasLocale } from "next-intl";
import { type Locale, routing } from "./routing";

/**
 * Narrows the `[locale]` param or 404s. Pages must not cast it: the layout's `hasLocale` guard renders in parallel
 * with the page, so `/favicon.ico` (→ locale "favicon.ico") used to reach catalog queries and throw a ZodError.
 */
export function asLocale(value: string): Locale {
	if (!hasLocale(routing.locales, value)) notFound();
	return value;
}
