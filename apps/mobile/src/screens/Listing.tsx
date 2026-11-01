import { useInfiniteQuery, useQuery } from "@tanstack/react-query";
import { type ReactElement, type ReactNode, useState } from "react";
import { ActivityIndicator, Pressable, Text, View } from "react-native";
import type { AppItemJSON } from "../../../../packages/app-client/src/api";
import type { ReadClient } from "../api/client";
import { type AppLocale, t } from "../i18n";
import { type CatalogType, SINGULAR } from "../navigation/catalog";
import { GENRE_METHOD, queryKeys } from "../query/keys";
import { formatScore } from "../ui/format";
import { type CatalogLabels, CatalogPage, type CatalogRow } from "../ui/screens/CatalogPage";

export type ListingReader = Pick<ReadClient, "listItems" | (typeof GENRE_METHOD)[CatalogType]>;

// Catalogue key suffix → server `sort` value (SDD 05 §6.2: `-` means descending).
export const SORTS = {
	popularity_desc: "-popularity",
	score_desc: "-score",
	release_date_desc: "-release_date",
	rating_count_desc: "-rating_count",
} as const;

type SortKey = keyof typeof SORTS;
type Filters = { sort: SortKey; genre: string | undefined };

const DEFAULTS: Filters = { sort: "popularity_desc", genre: undefined };

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

function Field({ label, children }: { label: string; children: ReactNode }): ReactElement {
	return (
		<View className="gap-2">
			<Text className="font-text text-14 font-semibold text-fg dark:text-fg-dark">{label}</Text>
			<View className="flex-row flex-wrap gap-2">{children}</View>
		</View>
	);
}

function toRow(item: AppItemJSON): CatalogRow {
	const year = Number.parseInt(item.releaseDate?.slice(0, 4) ?? "", 10);
	return {
		id: item.id,
		title: item.title ?? item.slug,
		year: Number.isNaN(year) ? null : year,
		// shortcut: list items carry no byline yet; upgrade when the API sends one.
		byline: null,
		ratingCount: item.ratingCount ?? 0,
		coverUri: item.coverThumbUrl ?? null,
		score: item.score ?? null,
	};
}

export function ListingBody({
	type,
	read,
	locale,
	onOpen,
}: {
	type: CatalogType;
	read: ListingReader;
	locale: AppLocale;
	onOpen?: (item: AppItemJSON) => void;
}): ReactElement {
	const [applied, setApplied] = useState<Filters>(DEFAULTS);
	const [draft, setDraft] = useState<Filters>(DEFAULTS);
	const [filtersOpen, setFiltersOpen] = useState(false);
	const [page, setPage] = useState(1);
	const filter = { sort: SORTS[applied.sort], ...(applied.genre ? { genre: applied.genre } : {}) };

	const genres = useQuery({
		queryKey: queryKeys.genres(type),
		queryFn: () => read[GENRE_METHOD[type]](),
	});
	// Pages are cached in order; page n needs page n-1's cursor, so Previous never refetches.
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

	const title = t(locale, `catalog.title.${SINGULAR[type]}`);
	// Pending also covers offline (paused query): keep the title so the screen is not blank.
	if (listing.isPending) {
		return (
			<View testID="listing" className="flex-1 gap-6 bg-bg px-4 pt-6 dark:bg-bg-dark">
				<Text
					accessibilityRole="header"
					className="font-display text-36 font-bold text-fg dark:text-fg-dark"
				>
					{title}
				</Text>
				<ActivityIndicator accessibilityLabel={t(locale, "item.loading")} />
			</View>
		);
	}

	const loaded = listing.data?.pages ?? [];
	const items = loaded[page - 1]?.data ?? [];
	const byId = new Map(items.map((item) => [item.id, item]));
	// The cursor API sends no total: the page count is known only once the last page is loaded.
	const pages = loaded.length + (listing.hasNextPage ? 1 : 0);

	const apply = (next: Filters): void => {
		setApplied(next);
		setDraft(next);
		setPage(1);
		setFiltersOpen(false);
	};
	const goTo = (next: number): void => {
		if (next > loaded.length) void listing.fetchNextPage();
		setPage(next);
	};

	const sortLabel = t(locale, `catalog.sorts.${applied.sort}`);
	const labels: CatalogLabels = {
		filters: (active) => t(locale, "catalog.filters", { count: active }),
		closeFilters: t(locale, "item.close"),
		sort: (current) => `${t(locale, "catalog.sort")}: ${current}`,
		found: (total) => t(locale, "catalog.resultCount", { count: total }),
		previous: t(locale, "catalog.previous"),
		next: t(locale, "catalog.next"),
		page: (current, count) =>
			listing.hasNextPage
				? t(locale, "catalog.page", { page: current })
				: t(locale, "item.pageOf", { page: current, count }),
		apply: t(locale, "catalog.apply"),
		reset: t(locale, "catalog.reset"),
		empty: t(locale, "catalog.empty"),
		meta: (row) =>
			[row.year, row.byline, t(locale, "catalog.ratingCount", { count: row.ratingCount })]
				.filter((part) => part !== null)
				.join(" · "),
		score: (row) =>
			row.score === null
				? t(locale, "score.none")
				: t(locale, "score.outOf", { score: formatScore(row.score, locale) }),
	};

	const filterPanel = (
		<>
			<Field label={t(locale, "catalog.sort")}>
				{(Object.keys(SORTS) as SortKey[]).map((key) => (
					<Chip
						key={key}
						label={t(locale, `catalog.sorts.${key}`)}
						selected={draft.sort === key}
						onPress={() => setDraft({ ...draft, sort: key })}
					/>
				))}
			</Field>
			<Field label={t(locale, "catalog.genre")}>
				<Chip
					label={t(locale, "catalog.anyGenre")}
					selected={draft.genre === undefined}
					onPress={() => setDraft({ ...draft, genre: undefined })}
				/>
				{/* shortcut: the genres endpoint sends slugs only; show a name once the API has one. */}
				{(genres.data?.data ?? []).map((g) => (
					<Chip
						key={g.slug}
						label={g.slug}
						selected={draft.genre === g.slug}
						onPress={() => setDraft({ ...draft, genre: g.slug })}
					/>
				))}
			</Field>
		</>
	);

	return (
		<CatalogPage
			title={title}
			rows={items.map(toRow)}
			total={null}
			page={page}
			pages={pages}
			activeFilters={applied.genre ? 1 : 0}
			sortLabel={sortLabel}
			filtersOpen={filtersOpen}
			filterPanel={filterPanel}
			labels={labels}
			locale={locale}
			onToggleFilters={() => {
				setDraft(applied);
				setFiltersOpen(!filtersOpen);
			}}
			onSort={() => {
				setDraft(applied);
				setFiltersOpen(true);
			}}
			onApply={() => apply(draft)}
			onReset={() => apply(DEFAULTS)}
			onPage={goTo}
			onOpen={(row) => {
				const item = byId.get(row.id);
				if (item) onOpen?.(item);
			}}
		/>
	);
}
