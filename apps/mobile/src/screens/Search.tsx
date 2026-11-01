import { useInfiniteQuery, useQueries } from "@tanstack/react-query";
import { type ReactElement, useEffect, useState } from "react";
import { Pressable, ScrollView, Text } from "react-native";
import type { AppItemJSON, SearchHit } from "../../../../packages/app-client/src/api";
import type { ReadClient } from "../api/client";
import { type AppLocale, t } from "../i18n";
import { CATALOG_TYPES, type CatalogType, itemHref, SINGULAR } from "../navigation/catalog";
import { queryKeys, TRENDING_METHOD } from "../query/keys";
import { formatScore } from "../ui/format";
import { type SearchLabels, type SearchRow, SearchSheet } from "../ui/screens/SearchSheet";

export type SearchReader = Pick<ReadClient, "searchIndex" | (typeof TRENDING_METHOD)[CatalogType]>;

// Same floor as the server's tooShort answer (SDD 07): shorter queries are never sent.
const MIN_CHARS = 2;
const RECENT_MAX = 5;
const TRENDING_MAX = 5;

const PLURAL = Object.fromEntries(CATALOG_TYPES.map((type) => [SINGULAR[type], type])) as Record<
	string,
	CatalogType
>;

function useDebounced(value: string, delay: number): string {
	const [settled, setSettled] = useState(value);
	useEffect(() => {
		const timer = setTimeout(() => setSettled(value), delay);
		return () => clearTimeout(timer);
	}, [value, delay]);
	return settled;
}

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
			className={`min-h-11 justify-center rounded-full border px-3 ${selected ? "border-fg bg-fg dark:border-fg-dark dark:bg-fg-dark" : "border-border dark:border-border-dark"}`}
			onPress={onPress}
		>
			<Text
				className={`font-text text-14 ${selected ? "text-bg dark:text-bg-dark" : "text-fg dark:text-fg-dark"}`}
			>
				{label}
			</Text>
		</Pressable>
	);
}

export function SearchBody({
	read,
	locale,
	onOpen,
	onCancel,
	delay = 300,
}: {
	read: SearchReader;
	locale: AppLocale;
	onOpen: (href: `/${CatalogType}/${string}`) => void;
	onCancel: () => void;
	delay?: number;
}): ReactElement {
	const [query, setQuery] = useState("");
	const [type, setType] = useState<CatalogType | undefined>(undefined);
	const [expanded, setExpanded] = useState(false);
	// shortcut: recent searches live for the session only; persist them once the app has a key-value store for prefs.
	const [recent, setRecent] = useState<string[]>([]);
	const q = useDebounced(query.trim(), delay);
	const enabled = q.length >= MIN_CHARS;

	const search = useInfiniteQuery({
		queryKey: ["searchIndex", { q, type, locale }] as const,
		enabled,
		initialPageParam: 1,
		queryFn: ({ pageParam }) =>
			read.searchIndex({
				query: { q, ...(type ? { type: SINGULAR[type] } : {}), locale, page: String(pageParam) },
			}),
		getNextPageParam: (last, all) => {
			const loaded = all.reduce((sum, page) => sum + (page.data?.length ?? 0), 0);
			return (last.data?.length ?? 0) > 0 && loaded < (last.total ?? 0)
				? all.length + 1
				: undefined;
		},
	});

	// Same keys as the home screen: the cached (and persisted) trending lists are reused.
	const trendingResults = useQueries({
		queries: CATALOG_TYPES.map((key) => ({
			queryKey: queryKeys.trending(key, locale),
			queryFn: () => read[TRENDING_METHOD[key]]({ query: { locale } }),
		})),
	});
	// Top items across types: lowest rank first, ties in catalogue order (sort is stable); failed types are skipped.
	const trending = trendingResults
		.flatMap((result) => result.data?.data ?? [])
		.filter((entry) => entry.item)
		.sort((a, b) => a.rank - b.rank)
		.slice(0, TRENDING_MAX)
		.map((entry) => entry.item as AppItemJSON);

	const remember = (text: string): void =>
		setRecent((list) => [text, ...list.filter((r) => r !== text)].slice(0, RECENT_MAX));

	const hits = (search.data?.pages ?? []).flatMap((page) => page.data ?? []);
	const byId = new Map<string, { type?: string; slug?: string }>([
		...trending.map((item) => [item.id, item] as const),
		...hits.map((hit) => [hit.id ?? "", hit] as const),
	]);
	const toRow = (hit: SearchHit): SearchRow => {
		const plural = PLURAL[hit.type ?? ""];
		return {
			id: hit.id ?? "",
			title: hit.title ?? hit.slug ?? "",
			year: hit.year ?? null,
			typeLabel: plural ? t(locale, `itemTypes.${SINGULAR[plural]}`) : "",
			coverUri: hit.cover?.thumb || null,
			score: hit.score ?? null,
		};
	};

	const trendingRow = (item: AppItemJSON): SearchRow => {
		const plural = PLURAL[item.type];
		const year = Number.parseInt(item.releaseDate?.slice(0, 4) ?? "", 10);
		return {
			id: item.id,
			title: item.title ?? item.slug,
			year: Number.isNaN(year) ? null : year,
			typeLabel: plural ? t(locale, `itemTypes.${SINGULAR[plural]}`) : "",
			coverUri: item.coverThumbUrl || null,
			score: item.score ?? null,
		};
	};

	const state = ((): Parameters<typeof SearchSheet>[0]["state"] => {
		if (!enabled) return { kind: "empty", recent, trending: trending.map(trendingRow) };
		if (search.isError) return { kind: "error", message: t(locale, "search.unavailable") };
		if (search.isPending) return { kind: "loading" };
		return { kind: "results", rows: hits.map(toRow), total: search.data.pages[0]?.total ?? 0 };
	})();

	const labels: SearchLabels = {
		placeholder: t(locale, "nav.search"),
		cancel: t(locale, "search.palette.close"),
		recent: t(locale, "search.palette.recent"),
		// shortcut: no "remove recent search" key yet (resources/lang is locked by P5.B.4.3); add one after it merges.
		removeRecent: (text) => t(locale, "search.palette.removeRecent", { q: text }),
		trending: t(locale, "home.trendingWeek"),
		trendingEmpty: t(locale, "search.palette.trendingEmpty"),
		loading: t(locale, "search.searching"),
		noResults: (text) => t(locale, "search.noResults", { q: text }),
		firstOf: (_shown, total) => t(locale, "search.count", { n: total }),
		seeAll: (text) =>
			expanded ? t(locale, "search.next") : t(locale, "search.palette.seeAll", { q: text }),
		score: (row) =>
			row.score === null
				? t(locale, "score.none")
				: t(locale, "score.outOf", { score: formatScore(row.score, locale) }),
	};

	const pick = (next: CatalogType | undefined): void => {
		setType(next);
		setExpanded(false);
	};
	const filters = (
		<ScrollView
			horizontal
			className="grow-0"
			contentContainerClassName="gap-2 px-4 py-2"
			keyboardShouldPersistTaps="handled"
		>
			<Chip
				label={t(locale, "search.all")}
				selected={type === undefined}
				onPress={() => pick(undefined)}
			/>
			{CATALOG_TYPES.map((key) => (
				<Chip
					key={key}
					label={t(locale, `nav.${key}`)}
					selected={type === key}
					onPress={() => pick(key)}
				/>
			))}
		</ScrollView>
	);

	const seeAll = (): void => {
		if (!enabled) return;
		remember(q);
		setExpanded(true);
	};

	return (
		<SearchSheet
			query={query}
			onQueryChange={(text) => {
				setQuery(text);
				setExpanded(false);
			}}
			onCancel={onCancel}
			onOpen={(row) => {
				const hit = byId.get(row.id);
				const plural = PLURAL[hit?.type ?? ""];
				if (!hit?.slug || !plural) return;
				if (enabled) remember(q);
				onOpen(itemHref(plural, { slug: hit.slug, id: row.id }));
			}}
			onRecent={(text) => {
				setQuery(text);
				setExpanded(false);
			}}
			onRemoveRecent={(text) => setRecent((list) => list.filter((r) => r !== text))}
			onSeeAll={
				expanded ? (search.hasNextPage ? () => void search.fetchNextPage() : undefined) : seeAll
			}
			state={state}
			labels={labels}
			locale={locale}
			limit={expanded ? Number.POSITIVE_INFINITY : undefined}
			filters={filters}
		/>
	);
}
