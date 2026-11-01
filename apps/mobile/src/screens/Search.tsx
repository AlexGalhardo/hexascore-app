import { useInfiniteQuery, useQueries } from "@tanstack/react-query";
import { type ReactElement, useState } from "react";
import { Pressable, ScrollView, Text } from "react-native";
import type { AppItemJSON, SearchHit, SuggestItem } from "../../../../packages/app-client/src/api";
import type { ReadClient } from "../api/client";
import { type AppLocale, t } from "../i18n";
import { CATALOG_TYPES, type CatalogType, itemHref, SINGULAR } from "../navigation/catalog";
import { queryKeys, TRENDING_METHOD } from "../query/keys";
import type { SearchClick } from "../search/events";
import { cutQuery, MIN_QUERY, useSuggest } from "../search/useSuggest";
import { formatScore } from "../ui/format";
import {
	type SearchGroup,
	type SearchLabels,
	type SearchRow,
	SearchSheet,
} from "../ui/screens/SearchSheet";

export type SearchReader = Pick<
	ReadClient,
	"searchIndex" | "searchSuggest" | (typeof TRENDING_METHOD)[CatalogType]
>;

const RECENT_MAX = 5;
const TRENDING_MAX = 5;

const PLURAL = Object.fromEntries(CATALOG_TYPES.map((type) => [SINGULAR[type], type])) as Record<
	string,
	CatalogType
>;

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

/**
 * Search tab (SDD 10 §6.1): debounced GET /search/suggest while typing; the search key or "see all" runs the
 * full GET /search, grouped by type; a click on a full-search result is reported through onClick.
 */
export function SearchBody({
	read,
	locale,
	onOpen,
	onCancel,
	onClick,
	delay = 300,
}: {
	read: SearchReader;
	locale: AppLocale;
	onOpen: (href: `/${CatalogType}/${string}`) => void;
	onCancel: () => void;
	/** POST /search/events for a click on a full-search result (SDD 07 §7.8). */
	onClick?: (click: SearchClick) => void;
	delay?: number;
}): ReactElement {
	const [query, setQuery] = useState("");
	const [type, setType] = useState<CatalogType | undefined>(undefined);
	// The query the full search ran for; null while the user is typing (suggestions).
	const [submitted, setSubmitted] = useState<string | null>(null);
	// shortcut: recent searches live for the session only; persist them once the app has a key-value store for prefs.
	const [recent, setRecent] = useState<string[]>([]);
	const singular = type ? SINGULAR[type] : undefined;
	const typed = cutQuery(query);
	const { q, suggest } = useSuggest(read, {
		query,
		type: singular,
		locale,
		delay,
		enabled: submitted === null,
	});

	const search = useInfiniteQuery({
		queryKey: ["searchIndex", { q: submitted, type, locale }] as const,
		enabled: submitted !== null,
		initialPageParam: 1,
		queryFn: ({ pageParam }) =>
			read.searchIndex({
				query: {
					q: submitted ?? "",
					...(singular ? { type: singular } : {}),
					locale,
					page: String(pageParam),
				},
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

	const typeLabel = (apiType: string | undefined): string => {
		const plural = PLURAL[apiType ?? ""];
		return plural ? t(locale, `itemTypes.${SINGULAR[plural]}`) : "";
	};

	// Each hit keeps its page's searchId and its 1-based position within its type group: that is what the
	// click event reports, never the query text.
	const clicks = new Map<string, SearchClick>();
	const grouped = new Map<string, SearchHit[]>();
	for (const page of search.data?.pages ?? []) {
		for (const hit of page.data ?? []) {
			if (!hit.id || !hit.type) continue;
			const list = grouped.get(hit.type) ?? [];
			list.push(hit);
			grouped.set(hit.type, list);
			clicks.set(hit.id, {
				searchId: page.searchId,
				refId: hit.id,
				position: list.length,
				group: hit.type,
			});
		}
	}
	const hits = [...grouped.values()].flat();
	// Entity suggestions (people, companies) have no item route here.
	const suggestions = (suggest.data?.data ?? []).filter((item) => PLURAL[item.type ?? ""]);

	const byId = new Map<string, { type?: string; slug?: string }>([
		...trending.map((item) => [item.id, item] as const),
		...suggestions.map(
			(item) => [item.id ?? "", { type: item.type, slug: item.path?.split("/").pop() }] as const,
		),
		...hits.map((hit) => [hit.id ?? "", hit] as const),
	]);
	const toRow = (hit: SearchHit): SearchRow => ({
		id: hit.id ?? "",
		title: hit.title ?? hit.slug ?? "",
		year: hit.year ?? null,
		typeLabel: typeLabel(hit.type),
		coverUri: hit.cover?.thumb || null,
		score: hit.score ?? null,
	});
	const suggestionRow = (item: SuggestItem): SearchRow => ({
		id: item.id ?? "",
		title: item.title ?? "",
		year: item.year ?? null,
		typeLabel: typeLabel(item.type),
		coverUri: item.coverThumbUrl || null,
	});
	const trendingRow = (item: AppItemJSON): SearchRow => {
		const year = Number.parseInt(item.releaseDate?.slice(0, 4) ?? "", 10);
		return {
			id: item.id,
			title: item.title ?? item.slug,
			year: Number.isNaN(year) ? null : year,
			typeLabel: typeLabel(item.type),
			coverUri: item.coverThumbUrl || null,
			score: item.score ?? null,
		};
	};
	const groups: SearchGroup[] = CATALOG_TYPES.flatMap((key) => {
		const list = grouped.get(SINGULAR[key]);
		return list ? [{ key, label: t(locale, `nav.${key}`), rows: list.map(toRow) }] : [];
	});

	const state = ((): Parameters<typeof SearchSheet>[0]["state"] => {
		if (submitted !== null) {
			if (search.isError) return { kind: "error", message: t(locale, "search.unavailable") };
			if (search.isPending) return { kind: "loading" };
			return { kind: "groups", groups, total: search.data.pages[0]?.total ?? 0 };
		}
		if (typed.length < MIN_QUERY || q.length < MIN_QUERY)
			return {
				kind: "empty",
				recent,
				trending: trending.map(trendingRow),
				// isLoading, not isPending: a paused (offline) query with no cache falls back to the empty message.
				trendingLoading: trendingResults.some((result) => result.isLoading),
			};
		if (suggest.isError) return { kind: "error", message: t(locale, "search.unavailable") };
		if (suggest.isPending) return { kind: "loading" };
		return {
			kind: "results",
			rows: suggestions.map(suggestionRow),
			total: suggestions.length,
			heading: t(locale, "search.suggestions"),
		};
	})();

	const labels: SearchLabels = {
		placeholder: t(locale, "nav.search"),
		cancel: t(locale, "search.palette.close"),
		recent: t(locale, "search.palette.recent"),
		// shortcut: no "remove recent search" key yet (resources/lang is locked by P5.B.4.3); add one after it merges.
		removeRecent: (text) => t(locale, "search.palette.removeRecent", { q: text }),
		trending: t(locale, "home.trendingWeek"),
		trendingEmpty: t(locale, "search.palette.trendingEmpty"),
		trendingLoading: t(locale, "item.loading"),
		loading: t(locale, "search.searching"),
		noResults: (text) => t(locale, "search.noResults", { q: text }),
		firstOf: (_shown, total) => t(locale, "search.count", { n: total }),
		seeAll: (text) =>
			submitted !== null
				? t(locale, "search.next")
				: t(locale, "search.palette.seeAll", { q: text }),
		score: (row) =>
			row.score === null || row.score === undefined
				? t(locale, "score.none")
				: t(locale, "score.outOf", { score: formatScore(row.score, locale) }),
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
				onPress={() => setType(undefined)}
			/>
			{CATALOG_TYPES.map((key) => (
				<Chip
					key={key}
					label={t(locale, `nav.${key}`)}
					selected={type === key}
					onPress={() => setType(key)}
				/>
			))}
		</ScrollView>
	);

	const runSearch = (): void => {
		if (typed.length < MIN_QUERY) return;
		remember(typed);
		setSubmitted(typed);
	};

	return (
		<SearchSheet
			query={query}
			onQueryChange={(text) => {
				setQuery(text);
				setSubmitted(null);
			}}
			onCancel={onCancel}
			onOpen={(row) => {
				const hit = byId.get(row.id);
				const plural = PLURAL[hit?.type ?? ""];
				if (!hit?.slug || !plural) return;
				if (typed.length >= MIN_QUERY) remember(typed);
				const click = submitted !== null ? clicks.get(row.id) : undefined;
				if (click) onClick?.(click);
				onOpen(itemHref(plural, { slug: hit.slug, id: row.id }));
			}}
			onRecent={(text) => {
				setQuery(text);
				setSubmitted(null);
			}}
			onRemoveRecent={(text) => setRecent((list) => list.filter((r) => r !== text))}
			onSeeAll={
				submitted !== null
					? search.hasNextPage
						? () => void search.fetchNextPage()
						: undefined
					: runSearch
			}
			onSubmit={runSearch}
			state={state}
			labels={labels}
			locale={locale}
			filters={filters}
		/>
	);
}
