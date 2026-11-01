import type { ReactElement } from "react";
import { Text } from "react-native";
import type { ItemJSON } from "../../../../../packages/app-client/src/api";
import { type AppLocale, t } from "../../i18n";

const ATTRIBUTED_SOURCES = new Set(["tmdb", "igdb"]);

/** True when any shown field came from TMDB or IGDB, whose terms require the notice (FR-10-9, open issue 01-6). */
export function needsAttribution(item: ItemJSON): boolean {
	return Object.values(item.provenance ?? {}).some((field) =>
		ATTRIBUTED_SOURCES.has(field.source ?? ""),
	);
}

export function Attribution({
	item,
	locale,
}: {
	item: ItemJSON;
	locale: AppLocale;
}): ReactElement | null {
	if (!needsAttribution(item)) return null;
	return (
		<Text testID="attribution" className="py-3 font-text text-12 text-muted dark:text-muted-dark">
			{t(locale, "item.attribution")}
		</Text>
	);
}
