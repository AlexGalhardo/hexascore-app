import { usePage } from "@inertiajs/react";

import { type I18nProps, translate } from "./i18n";

// useT returns the translator for the page's locale; the server shares only
// the catalogue namespaces listed in routes/web.go (inertia.I18n).
export function useT(): (key: string, params?: Record<string, string | number>) => string {
	const { locale, messages } = usePage<I18nProps>().props;
	return (key, params) => translate(messages, locale, key, params);
}
