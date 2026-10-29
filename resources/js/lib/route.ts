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
