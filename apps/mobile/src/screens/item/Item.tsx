import { useQuery, useQueryClient } from "@tanstack/react-query";
import { type ReactElement, useEffect, useRef, useState } from "react";
import { Pressable, ScrollView, Text, View } from "react-native";
import type { EntityRef } from "../../../../../packages/app-client/src/api";
import type { ReadClient } from "../../api/client";
import { formatScore } from "../../components/ItemRow";
import { type AppLocale, t } from "../../i18n";
import type { CatalogType } from "../../navigation/catalog";
import { queryKeys } from "../../query/keys";
import { Attribution } from "./Attribution";
import { ItemHeader } from "./ItemHeader";

export type ItemReader = Pick<ReadClient, "getItem" | "listEditions" | "me">;

/** Mean of the user's stars per category from GET /items/{id}/me, which the generated client types `unknown`. */
export function overallOf(me: unknown): number | null {
	if (typeof me !== "object" || me === null) return null;
	const rating = (me as { rating?: unknown }).rating;
	if (typeof rating !== "object" || rating === null) return null;
	const stars = Object.values(rating).filter((value): value is number => typeof value === "number");
	if (stars.length === 0) return null;
	return stars.reduce((sum, value) => sum + value, 0) / stars.length;
}

function groupByRole(entities: EntityRef[]): [string, EntityRef[]][] {
	const groups = new Map<string, EntityRef[]>();
	for (const entity of entities) {
		const role = entity.role ?? "";
		groups.set(role, [...(groups.get(role) ?? []), entity]);
	}
	return [...groups];
}

export function ItemView({
	type,
	id: requestedId,
	read,
	locale,
	onMerged,
	onOpenEntity,
}: {
	type: CatalogType;
	id: string;
	read: ItemReader;
	locale: AppLocale;
	/** Called with the surviving id after a merged id was followed (301), so the route can be rewritten. */
	onMerged?: (id: string) => void;
	onOpenEntity?: (id: string) => void;
}): ReactElement {
	const queryClient = useQueryClient();
	const [id, setId] = useState(requestedId);
	const mergedFrom = useRef<string | null>(null);

	const item = useQuery({
		queryKey: queryKeys.item(type, id, locale),
		queryFn: () => read.getItem({ params: { type, id }, query: { locale } }),
	});

	// fetch follows the 301, so a merged id answers with the surviving item: move it to that item's key.
	const data = item.data;
	useEffect(() => {
		if (!data || data.id === id) return;
		queryClient.setQueryData(queryKeys.item(type, data.id, locale), data);
		mergedFrom.current = id;
		setId(data.id);
		onMerged?.(data.id);
	}, [data, id, type, locale, queryClient, onMerged]);
	// Removed only once the observer is on the new key; removing it earlier would refetch the old id.
	useEffect(() => {
		if (mergedFrom.current === null || mergedFrom.current === id) return;
		queryClient.removeQueries({
			queryKey: queryKeys.item(type, mergedFrom.current, locale),
			exact: true,
		});
		mergedFrom.current = null;
	}, [id, type, locale, queryClient]);

	const resolved = data && data.id === id ? data : undefined;
	const editions = useQuery({
		queryKey: queryKeys.editions(id),
		queryFn: () => read.listEditions({ params: { id } }),
		enabled: resolved !== undefined && type === "books",
	});
	const me = useQuery({
		queryKey: queryKeys.itemMe(id),
		queryFn: () => read.me({ params: { id } }),
		enabled: resolved !== undefined,
	});

	if (!resolved) {
		return (
			<View testID="item" className="flex-1 items-center justify-center bg-bg px-4 dark:bg-bg-dark">
				{item.isError ? (
					<Text className="text-16 text-fg dark:text-fg-dark">
						{t(locale, "errors.notFoundTitle")}
					</Text>
				) : null}
			</View>
		);
	}

	const overall = overallOf(me.data);
	const editionRows = editions.data?.data ?? [];
	return (
		<ScrollView testID="item" className="flex-1 bg-bg dark:bg-bg-dark">
			<ItemHeader item={resolved} locale={locale} />
			{overall !== null ? (
				<View className="flex-row justify-between px-4 pt-3">
					<Text className="text-16 text-fg dark:text-fg-dark">
						{t(locale, "ratings.yourOverall")}
					</Text>
					<Text className="text-16 text-fg dark:text-fg-dark">{formatScore(overall, locale)}</Text>
				</View>
			) : null}
			{groupByRole(resolved.entities ?? []).map(([role, entities]) => (
				<View key={role} className="gap-1 px-4 pt-3">
					<Text className="text-16 text-fg dark:text-fg-dark">
						{t(locale, `item.roles.${role}`)}
					</Text>
					{entities.map((entity) => (
						<Pressable
							key={entity.id ?? entity.name}
							accessibilityRole="link"
							className="min-h-11 justify-center"
							onPress={() => {
								if (entity.id) onOpenEntity?.(entity.id);
							}}
						>
							<Text className="text-14 text-fg dark:text-fg-dark">{entity.name}</Text>
						</Pressable>
					))}
				</View>
			))}
			{editionRows.length > 0 ? (
				<View className="gap-1 px-4 pt-3">
					<Text className="text-16 text-fg dark:text-fg-dark">
						{t(locale, "item.editions.title")}
					</Text>
					{editionRows.map((edition) => (
						<Text key={edition.id} className="text-14 text-fg dark:text-fg-dark">
							{edition.title ?? edition.isbn13 ?? edition.id}
						</Text>
					))}
				</View>
			) : null}
			<Attribution item={resolved} locale={locale} />
		</ScrollView>
	);
}
