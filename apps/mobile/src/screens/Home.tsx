import { useQueries } from "@tanstack/react-query";
import type { ReactElement } from "react";
import { ScrollView, Text, View } from "react-native";
import type { ItemJSON, TrendingEntryJSON } from "../../../../packages/app-client/src/api";
import type { ReadClient } from "../api/client";
import { type AppLocale, t } from "../i18n";
import { CATALOG_TYPES, type CatalogType, SINGULAR } from "../navigation/catalog";
import { queryKeys, TRENDING_METHOD } from "../query/keys";
import { formatScore } from "../ui/format";
import {
	GRID,
	type HomeItem,
	type HomeLabels,
	type HomeSection,
	HomeSections,
	type ItemType,
	TRENDING,
} from "../ui/screens/HomeSections";

export type TrendingReader = Pick<ReadClient, (typeof TRENDING_METHOD)[CatalogType]>;

const PLURAL = Object.fromEntries(CATALOG_TYPES.map((type) => [SINGULAR[type], type])) as Record<
	ItemType,
	CatalogType
>;

// SDD 08 §6: an all-prior list is titled "Popular {type}".
function trendingTitle(type: CatalogType, entries: TrendingEntryJSON[], locale: AppLocale): string {
	const allPrior = entries.every((entry) => entry.basis === "prior");
	return t(locale, allPrior ? "catalog.trending.allPrior" : "catalog.trending.title", {
		type: t(locale, `catalog.title.${SINGULAR[type]}`),
	});
}

function homeItem(item: ItemJSON): HomeItem {
	const year = Number.parseInt(item.releaseDate?.slice(0, 4) ?? "", 10);
	return {
		id: item.id,
		title: item.title ?? item.slug,
		year: Number.isNaN(year) ? null : year,
		// shortcut: /app/v1 items carry no cover URL yet (FIX-app-item-cover-urls); Cover reserves the 2:3 box.
		coverUri: null,
		score: item.score ?? null,
		ratingCount: item.ratingCount ?? 0,
	};
}

export function HomeBody({
	read,
	locale,
	onOpen,
	onSeeAll,
}: {
	read: TrendingReader;
	locale: AppLocale;
	onOpen?: (item: ItemJSON) => void;
	onSeeAll?: (type: CatalogType) => void;
}): ReactElement {
	const results = useQueries({
		queries: CATALOG_TYPES.map((type) => ({
			queryKey: queryKeys.trending(type, locale),
			queryFn: () => read[TRENDING_METHOD[type]]({ query: { locale } }),
		})),
	});
	const settled = results.every((result) => !result.isPending);

	// At most 4 × 10 entries: rebuilt per render, no memo.
	const sources = new Map<string, { item: ItemJSON; basis: TrendingEntryJSON["basis"] }>();
	const sections: HomeSection[] = [];
	CATALOG_TYPES.forEach((type, index) => {
		// A failed or empty type is hidden (ROADMAP P5.B.1.1); nothing is ever filled in client-side.
		const entries = (results[index]?.data?.data ?? []).filter((entry) => entry.item);
		if (entries.length === 0) return;
		const items = entries.map((entry) => {
			const item = entry.item as ItemJSON;
			sources.set(item.id, { item, basis: entry.basis });
			return homeItem(item);
		});
		sections.push({
			type: SINGULAR[type] as ItemType,
			title: trendingTitle(type, entries, locale),
			typeLabel: t(locale, `catalog.title.${SINGULAR[type]}`),
			trending: items.slice(0, TRENDING),
			grid: items.slice(TRENDING, TRENDING + GRID),
		});
	});

	const labels: HomeLabels = {
		seeAll: (section) => t(locale, "catalog.landing.seeAll", { section: section.typeLabel }),
		// SDD 08 §6: a prior entry is labelled Popular instead of trending.
		rank: (position, item) =>
			`#${position} · ${t(locale, sources.get(item.id)?.basis === "prior" ? "home.popular" : "home.trendingWeek")}`,
		verdict: (item) => t(locale, item.score === null ? "item.notEnough" : "item.communityScore"),
		ratings: (item) => t(locale, "catalog.ratingCount", { count: item.ratingCount }),
		score: (item) =>
			item.score === null
				? t(locale, "score.none")
				: t(locale, "score.outOf", { score: formatScore(item.score, locale) }),
	};

	return (
		<ScrollView className="flex-1 bg-bg dark:bg-bg-dark">
			<View className="px-4 pt-4">
				<Text
					accessibilityRole="header"
					className="font-display text-24 font-bold text-fg dark:text-fg-dark"
				>
					HexaScore
				</Text>
			</View>
			{sections.length === 0 && settled ? (
				<Text className="px-4 py-8 text-16 text-muted dark:text-muted-dark">
					{t(locale, "catalog.trending.empty")}
				</Text>
			) : (
				<HomeSections
					sections={sections}
					labels={labels}
					locale={locale}
					onOpen={(item) => {
						const source = sources.get(item.id);
						if (source) onOpen?.(source.item);
					}}
					onSeeAll={(type) => onSeeAll?.(PLURAL[type])}
				/>
			)}
		</ScrollView>
	);
}
