import { useInfiniteQuery, useQuery } from "@tanstack/react-query";
import { type ReactElement, useCallback, useState } from "react";
import { FlatList, Pressable, ScrollView, Text, View } from "react-native";
import type { ItemJSON } from "../../../../packages/app-client/src/api";
import type { ReadClient } from "../api/client";
import { ITEM_ROW_HEIGHT, ItemRow } from "../components/ItemRow";
import { type AppLocale, t } from "../i18n";
import { type CatalogType, SINGULAR } from "../navigation/catalog";
import { GENRE_METHOD, queryKeys } from "../query/keys";

export type ListingReader = Pick<ReadClient, "listItems" | (typeof GENRE_METHOD)[CatalogType]>;

// Catalogue key suffix → server `sort` value (SDD 05 §6.2: `-` means descending).
export const SORTS = {
	popularity_desc: "-popularity",
	score_desc: "-score",
	release_date_desc: "-release_date",
	rating_count_desc: "-rating_count",
} as const;

type SortKey = keyof typeof SORTS;

function Chip({
	label,
	selected,
	onPress,
}: {
	label: string;
	selected: boolean;
	onPress: () => void;
}): ReactElement {
	return (
		<Pressable
			accessibilityRole="button"
			accessibilityState={{ selected }}
			className={`min-h-11 justify-center rounded-full border px-3 ${selected ? "border-fg dark:border-fg-dark" : "border-border dark:border-border-dark"}`}
			onPress={onPress}
		>
			<Text className="text-14 text-fg dark:text-fg-dark">{label}</Text>
		</Pressable>
	);
}

const getItemLayout = (_: unknown, index: number) => ({
	length: ITEM_ROW_HEIGHT,
	offset: ITEM_ROW_HEIGHT * index,
	index,
});

export function ListingBody({
	type,
	read,
	locale,
	onOpen,
}: {
	type: CatalogType;
	read: ListingReader;
	locale: AppLocale;
	onOpen?: (item: ItemJSON) => void;
}): ReactElement {
	const [sort, setSort] = useState<SortKey>("popularity_desc");
	const [genre, setGenre] = useState<string | undefined>(undefined);
	const filter = { sort: SORTS[sort], ...(genre ? { genre } : {}) };

	const genres = useQuery({
		queryKey: queryKeys.genres(type),
		queryFn: () => read[GENRE_METHOD[type]](),
	});
	const listing = useInfiniteQuery({
		queryKey: queryKeys.listing(type, filter, locale),
		initialPageParam: undefined as string | undefined,
		queryFn: ({ pageParam }) =>
			read.listItems({
				params: { type },
				query: { ...filter, locale, ...(pageParam ? { cursor: pageParam } : {}) },
			}),
		// Opaque, server-signed cursor (VF-005): passed back as received, never built here.
		getNextPageParam: (last) => last.meta?.next_cursor || undefined,
	});
	const rows = listing.data?.pages.flatMap((page) => page.data ?? []) ?? [];

	const { hasNextPage, isFetchingNextPage, fetchNextPage } = listing;
	const onEndReached = useCallback(() => {
		if (hasNextPage && !isFetchingNextPage) void fetchNextPage();
	}, [hasNextPage, isFetchingNextPage, fetchNextPage]);
	const renderItem = useCallback(
		({ item }: { item: ItemJSON }) => <ItemRow item={item} locale={locale} onPress={onOpen} />,
		[locale, onOpen],
	);

	return (
		<FlatList
			testID="listing"
			className="flex-1 bg-bg dark:bg-bg-dark"
			data={rows}
			keyExtractor={(item) => item.id}
			renderItem={renderItem}
			getItemLayout={getItemLayout}
			windowSize={7}
			maxToRenderPerBatch={10}
			removeClippedSubviews
			onEndReached={onEndReached}
			onEndReachedThreshold={0.5}
			ListHeaderComponent={
				<View className="gap-2 px-4 pt-4">
					<Text accessibilityRole="header" className="text-24 text-fg dark:text-fg-dark">
						{t(locale, `catalog.title.${SINGULAR[type]}`)}
					</Text>
					<ScrollView horizontal contentContainerClassName="gap-2">
						{(Object.keys(SORTS) as SortKey[]).map((key) => (
							<Chip
								key={key}
								label={t(locale, `catalog.sorts.${key}`)}
								selected={sort === key}
								onPress={() => setSort(key)}
							/>
						))}
					</ScrollView>
					<ScrollView horizontal contentContainerClassName="gap-2">
						<Chip
							label={t(locale, "catalog.anyGenre")}
							selected={genre === undefined}
							onPress={() => setGenre(undefined)}
						/>
						{(genres.data?.data ?? []).map((g) => (
							<Chip
								key={g.slug}
								label={g.slug}
								selected={genre === g.slug}
								onPress={() => setGenre(g.slug)}
							/>
						))}
					</ScrollView>
				</View>
			}
			ListEmptyComponent={
				listing.isPending ? undefined : (
					<Text className="px-4 py-8 text-16 text-muted dark:text-muted-dark">
						{t(locale, "catalog.empty")}
					</Text>
				)
			}
		/>
	);
}
