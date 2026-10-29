// Locale-prefixed URLs (SDD 09 §7.1). Wraps the path a `vukem codegen:routes`
// helper returns; `en` is the unprefixed default.
export const LOCALE_PREFIXES: Readonly<Record<string, string>> = {
	en: "",
	"pt-BR": "/pt-br",
	es: "/es",
};

export function localized(path: string, locale: string): string {
	const prefix = LOCALE_PREFIXES[locale];
	// Only same-origin absolute paths are prefixed; "//host" is protocol-relative.
	if (!prefix || !path.startsWith("/") || path.startsWith("//")) return path;
	return prefix + path;
}

// unlocalized strips a locale prefix, for building the same page in another locale.
export function unlocalized(path: string): string {
	for (const prefix of Object.values(LOCALE_PREFIXES)) {
		if (
			prefix &&
			(path === prefix || path.startsWith(`${prefix}/`) || path.startsWith(`${prefix}?`))
		) {
			return path.slice(prefix.length) || "/";
		}
	}
	return path;
}
