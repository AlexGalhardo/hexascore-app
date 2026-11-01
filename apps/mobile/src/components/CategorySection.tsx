import type { ReactElement } from "react";
import { Text, View } from "react-native";
import type { ItemJSON, TrendingEntryJSON } from "../../../../packages/app-client/src/api";
import { type AppLocale, t } from "../i18n";
import { type CatalogType, SINGULAR } from "../navigation/catalog";
import { ItemRow } from "./ItemRow";

// SDD 08 §6: prior entries get the chip; an all-prior list is titled "Popular {type}".
export function trendingTitle(
	type: CatalogType,
	entries: TrendingEntryJSON[],
	locale: AppLocale,
): string {
	const label = t(locale, `catalog.title.${SINGULAR[type]}`);
	const allPrior = entries.length > 0 && entries.every((entry) => entry.basis === "prior");
	return t(locale, allPrior ? "catalog.trending.allPrior" : "catalog.trending.title", {
		type: label,
	});
}

export function CategorySection({
	type,
	entries,
	locale,
	onOpen,
}: {
	type: CatalogType;
	entries: TrendingEntryJSON[];
	locale: AppLocale;
	onOpen?: (item: ItemJSON) => void;
}): ReactElement {
	return (
		<View className="gap-1 py-3">
			<Text
				testID="section-title"
				accessibilityRole="header"
				className="px-4 text-20 text-fg dark:text-fg-dark"
			>
				{trendingTitle(type, entries, locale)}
			</Text>
			{/* ≤ 10 ranks per type (FR-08-4): plain rows inside the home FlatList, no nested virtualisation. */}
			{entries.map((entry) =>
				entry.item ? (
					<ItemRow
						key={entry.item.id}
						item={entry.item}
						basis={entry.basis}
						locale={locale}
						onPress={onOpen}
					/>
				) : null,
			)}
		</View>
	);
}
