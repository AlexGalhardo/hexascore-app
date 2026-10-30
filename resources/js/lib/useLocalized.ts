import { usePage } from "@inertiajs/react";

import type { I18nProps } from "./i18n";
import { localized } from "./route";

// useLocalized prefixes an absolute path with the page's locale, so form posts reach the
// server under the same prefix and its redirects keep the language (SDD 09 §7.1).
export function useLocalized(): (path: string) => string {
	const { locale } = usePage<I18nProps>().props;
	return (path) => localized(path, locale);
}
