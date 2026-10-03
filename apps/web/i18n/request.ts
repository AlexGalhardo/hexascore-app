import { hasLocale } from "next-intl";
import { getRequestConfig } from "next-intl/server";
import { routing } from "./routing";

// One file per namespace (messages/<locale>/<namespace>.json) keeps parallel work conflict-free.
// Add new namespaces here; a missing file fails the build instead of rendering raw keys.
export const NAMESPACES = [
	"common",
	"errors",
	"billing",
	"categories",
	"catalog",
	"item",
	"search",
	"auth",
	"account",
	"contact",
] as const;

export default getRequestConfig(async ({ requestLocale }) => {
	const requested = await requestLocale;
	const locale = hasLocale(routing.locales, requested) ? requested : routing.defaultLocale;
	const parts = await Promise.all(
		NAMESPACES.map(
			async (ns) => (await import(`../messages/${locale}/${ns}.json`)).default as Record<string, unknown>,
		),
	);
	return { locale, messages: Object.assign({}, ...parts) };
});
