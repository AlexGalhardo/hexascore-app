import { usePage } from "@inertiajs/react";

import type { SharedProps } from "../types";
import { localized } from "./route";

// Prefixes in-app links with the visitor's locale so navigation keeps the language.
export function useLocalized(): (path: string) => string {
	const { locale } = usePage<SharedProps>().props;
	return (path) => localized(path, locale);
}
