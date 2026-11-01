import { useQueries } from "@tanstack/react-query";
import { type ReactElement, useCallback } from "react";
import { FlatList, Text, View } from "react-native";
import type { ItemJSON, TrendingEntryJSON } from "../../../../packages/app-client/src/api";
import type { ReadClient } from "../api/client";
import { CategorySection } from "../components/CategorySection";
import { type AppLocale, t } from "../i18n";
import { CATALOG_TYPES, type CatalogType } from "../navigation/catalog";
import { queryKeys, TRENDING_METHOD } from "../query/keys";

export type TrendingReader = Pick<ReadClient, (typeof TRENDING_METHOD)[CatalogType]>;

type Section = { type: CatalogType; entries: TrendingEntryJSON[] };

export function HomeBody({
	read,
	locale,
	onOpen,
}: {
	read: TrendingReader;
	locale: AppLocale;
	onOpen?: (item: ItemJSON) => void;
}): ReactElement {
	const results = useQueries({
		queries: CATALOG_TYPES.map((type) => ({
			queryKey: queryKeys.trending(type, locale),
			queryFn: () => read[TRENDING_METHOD[type]]({ query: { locale } }),
		})),
	});
	// A failed or empty type is hidden (ROADMAP P5.B.1.1); nothing is ever filled in client-side.
	const sections: Section[] = CATALOG_TYPES.map((type, index) => ({
		type,
		entries: results[index]?.data?.data ?? [],
	})).filter((section) => section.entries.length > 0);
	const settled = results.every((result) => !result.isPending);

	const renderItem = useCallback(
		({ item }: { item: Section }) => (
			<CategorySection type={item.type} entries={item.entries} locale={locale} onOpen={onOpen} />
		),
		[locale, onOpen],
	);

	return (
		<FlatList
			className="flex-1 bg-bg dark:bg-bg-dark"
			data={sections}
			keyExtractor={(section) => section.type}
			renderItem={renderItem}
			// Four sections of at most ten rows: all mount at once; windowing only matters for the listing.
			initialNumToRender={CATALOG_TYPES.length}
			ListHeaderComponent={
				<View className="px-4 pt-4">
					<Text accessibilityRole="header" className="text-24 text-fg dark:text-fg-dark">
						HexaScore
					</Text>
				</View>
			}
			ListEmptyComponent={
				settled ? (
					<Text className="px-4 py-8 text-16 text-muted dark:text-muted-dark">
						{t(locale, "catalog.trending.empty")}
					</Text>
				) : undefined
			}
		/>
	);
}
