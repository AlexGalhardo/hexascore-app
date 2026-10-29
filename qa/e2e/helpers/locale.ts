// The default locale (en) has no URL prefix; pt-BR lives under /pt-br and es under /es (SDD 09 §9).
export function localePrefix(locale: string): string {
	return locale === "en" ? "" : `/${locale.toLowerCase()}`;
}
