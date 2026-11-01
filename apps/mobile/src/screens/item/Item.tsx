import { useQuery, useQueryClient } from "@tanstack/react-query";
import { type ReactElement, useEffect, useRef, useState } from "react";
import { Modal, Pressable, Text, View } from "react-native";
import type { EntityRef, ItemJSON } from "../../../../../packages/app-client/src/api";
import type { ReadClient } from "../../api/client";
import { type AppLocale, t } from "../../i18n";
import { type CatalogType, SINGULAR } from "../../navigation/catalog";
import { queryKeys } from "../../query/keys";
import type { Scores } from "../../rating/cache";
import { RateSheet } from "../../rating/RateSheet";
import { formatScore } from "../../ui/format";
import { EditionsList } from "../../ui/screens/EditionsList";
import {
	ItemDetail,
	type ItemDetailLabels,
	type ItemDetailView,
} from "../../ui/screens/ItemDetail";
import { Attribution } from "./Attribution";

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

function ratingOf(me: unknown): Scores | null {
	const rating = (me as { rating?: unknown } | null | undefined)?.rating;
	return typeof rating === "object" && rating !== null && Object.keys(rating).length > 0
		? (rating as Scores)
		: null;
}

function groupByRole(entities: EntityRef[]): [string, EntityRef[]][] {
	const groups = new Map<string, EntityRef[]>();
	for (const entity of entities) {
		const role = entity.role ?? "";
		groups.set(role, [...(groups.get(role) ?? []), entity]);
	}
	return [...groups];
}

const H2 = "font-display text-24 font-bold text-fg dark:text-fg-dark";

function detailOf(item: ItemJSON, type: CatalogType, locale: AppLocale): ItemDetailView {
	const year = item.releaseDate?.slice(0, 4);
	return {
		title: item.title ?? item.slug,
		kicker: [t(locale, `catalog.title.${SINGULAR[type]}`), year].filter(Boolean).join(" · "),
		byline: null,
		// shortcut: /app/v1 items carry no cover URL yet (FIX-app-item-cover-urls); pass item.coverUrl once it lands.
		coverUri: null,
		score: item.score ?? null,
		ratingCount: item.ratingCount ?? 0,
		genres: (item.entities ?? []).filter((e) => e.role === "genre").map((e) => e.name),
		synopsis: item.synopsis?.text ?? null,
		categories: Object.entries(item.scoreByCategory ?? {}).map(([key, score]) => ({
			key,
			label: t(locale, `categories.${SINGULAR[type]}.${key}`),
			score,
		})),
	};
}

function labelsOf(locale: AppLocale, rated: boolean): ItemDetailLabels {
	const score = (value: number | null): string =>
		value === null
			? t(locale, "score.none")
			: t(locale, "score.outOf", { score: formatScore(value, locale) });
	return {
		verdict: (value) => t(locale, value === null ? "item.notEnough" : "item.communityScore"),
		ratings: (count) => t(locale, "catalog.ratingCount", { count }),
		score,
		rate: t(locale, rated ? "ratings.update" : "item.rateIt"),
		addToList: t(locale, "lists.addToList"),
		categoryAverages: t(locale, "item.categoryScores"),
		synopsis: t(locale, "item.synopsis"),
		noSynopsis: t(locale, "item.noSynopsis"),
		// The inline panel is unused here (rating = null): rating goes through RateSheet.
		rateTitle: () => t(locale, "ratings.title"),
		overall: t(locale, "ratings.yourOverall"),
		save: t(locale, "ratings.save"),
		moreLikeThis: "",
	};
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
	const [rating, setRating] = useState(false);

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
	const groups = groupByRole((resolved.entities ?? []).filter((e) => e.role !== "genre"));
	const noop = (): void => {};
	return (
		<>
			<ItemDetail
				item={detailOf(resolved, type, locale)}
				similar={[]}
				rating={null}
				ownScore={
					overall === null ? null : { label: t(locale, "ratings.yourOverall"), score: overall }
				}
				labels={labelsOf(locale, overall !== null)}
				locale={locale}
				onRate={() => setRating(true)}
				onRateCategory={noop}
				onSave={noop}
				onOpenSimilar={noop}
			>
				{groups.map(([role, entities]) => (
					<View key={role} className="gap-1">
						<Text accessibilityRole="header" className={H2}>
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
								<Text className="font-text text-16 text-fg underline dark:text-fg-dark">
									{entity.name}
								</Text>
							</Pressable>
						))}
					</View>
				))}
				<EditionsList
					title={t(locale, "item.editions.title")}
					editions={editionRows.map((edition) => ({
						id: edition.id,
						title: edition.title ?? edition.isbn13 ?? edition.id,
						facts: edition.publisher
							? [{ label: t(locale, "item.roles.publisher"), value: edition.publisher }]
							: [],
					}))}
				/>
				<Attribution item={resolved} locale={locale} />
			</ItemDetail>
			<Modal visible={rating} animationType="slide" onRequestClose={() => setRating(false)}>
				<RateSheet
					itemId={id}
					type={type}
					initial={ratingOf(me.data)}
					locale={locale}
					onDone={() => setRating(false)}
				/>
			</Modal>
		</>
	);
}
