import type { ReactElement, ReactNode } from "react";
import { Pressable, ScrollView, Text, View } from "react-native";
import { Btn } from "../Btn";
import { Cover } from "../Cover";
import { Icon } from "../Icon";
import { ScoreBadge } from "../ScoreBadge";

export interface CatalogRow {
	id: string;
	title: string;
	year: number | null;
	/** Studio, director, creator or author, whatever the type shows. */
	byline: string | null;
	ratingCount: number;
	coverUri: string | null;
	score: number | null;
}

export interface CatalogLabels {
	filters: (active: number) => string;
	closeFilters: string;
	sort: (current: string) => string;
	found: (total: number) => string;
	previous: string;
	next: string;
	page: (page: number, pages: number) => string;
	apply: string;
	reset: string;
	empty: string;
	meta: (row: CatalogRow) => string;
	score: (row: CatalogRow) => string;
}

export interface CatalogPageProps {
	title: string;
	rows: CatalogRow[];
	/** Null when the API pages by cursor and sends no total: the count line is hidden. */
	total: number | null;
	page: number;
	pages: number;
	activeFilters: number;
	sortLabel: string;
	filtersOpen: boolean;
	/** Genre, platform, year, minimum score fields (pickers come from the wiring row). */
	filterPanel: ReactNode;
	labels: CatalogLabels;
	locale: string;
	onToggleFilters: () => void;
	onSort: () => void;
	onApply: () => void;
	onReset: () => void;
	onPage: (page: number) => void;
	onOpen: (row: CatalogRow) => void;
}

const MUTED = "font-text text-14 text-muted dark:text-muted-dark";

function Pager({
	page,
	pages,
	labels,
	onPage,
}: Pick<CatalogPageProps, "page" | "pages" | "labels" | "onPage">): ReactElement {
	return (
		<View className="flex-row items-center justify-between gap-2 py-3">
			<Btn
				variant="secondary"
				label={labels.previous}
				disabled={page <= 1}
				onPress={() => onPage(page - 1)}
			/>
			<Text className="font-text text-14 font-semibold tabular-nums text-fg dark:text-fg-dark">
				{labels.page(page, pages)}
			</Text>
			<Btn
				variant="secondary"
				label={labels.next}
				disabled={page >= pages}
				onPress={() => onPage(page + 1)}
			/>
		</View>
	);
}

function Row({
	row,
	labels,
	locale,
	onOpen,
}: {
	row: CatalogRow;
	labels: CatalogLabels;
	locale: string;
	onOpen: (row: CatalogRow) => void;
}): ReactElement {
	return (
		<Pressable
			accessibilityRole="button"
			accessibilityLabel={[row.title, labels.meta(row), labels.score(row)].join(", ")}
			onPress={() => onOpen(row)}
			className="min-h-11 flex-row items-center gap-4 border-b border-border py-3 active:bg-surface dark:border-border-dark dark:active:bg-surface-dark"
		>
			<Cover uri={row.coverUri} title={row.title} className="w-12" />
			<View className="flex-1">
				<Text className="font-display text-18 font-bold text-fg dark:text-fg-dark">
					{row.title}
				</Text>
				<Text className={MUTED}>{labels.meta(row)}</Text>
			</View>
			<ScoreBadge score={row.score} locale={locale} label={labels.score(row)} />
		</Pressable>
	);
}

// Round 2 catalogue index, mobile: filters push the list down instead of covering it.
export function CatalogPage(props: CatalogPageProps): ReactElement {
	const { title, rows, total, page, pages, activeFilters, sortLabel, filtersOpen, labels } = props;
	const toggle = filtersOpen
		? "bg-fg border-fg dark:bg-fg-dark dark:border-fg-dark"
		: "bg-field-bg border-fg dark:bg-field-bg-dark dark:border-fg-dark";
	const toggleText = filtersOpen ? "text-bg dark:text-bg-dark" : "text-fg dark:text-fg-dark";
	return (
		<ScrollView
			testID="listing"
			className="flex-1 bg-bg dark:bg-bg-dark"
			contentContainerClassName="px-4 pb-8"
		>
			<Text
				accessibilityRole="header"
				className="pb-4 pt-6 font-display text-36 font-bold text-fg dark:text-fg-dark"
			>
				{title}
			</Text>
			<View className="flex-row gap-2 border-y border-fg py-3 dark:border-fg-dark">
				<Pressable
					accessibilityRole="button"
					accessibilityLabel={filtersOpen ? labels.closeFilters : labels.filters(activeFilters)}
					accessibilityState={{ expanded: filtersOpen }}
					onPress={props.onToggleFilters}
					className={`min-h-11 flex-row items-center gap-2 rounded-sm border px-3 ${toggle}`}
				>
					<Icon name={filtersOpen ? "close" : "options-outline"} size={18} className={toggleText} />
					<Text className={`font-text text-16 font-semibold ${toggleText}`}>
						{labels.filters(activeFilters)}
					</Text>
				</Pressable>
				<Pressable
					accessibilityRole="button"
					accessibilityLabel={labels.sort(sortLabel)}
					onPress={props.onSort}
					className="min-h-11 flex-1 flex-row items-center justify-between rounded-sm border border-field bg-field-bg px-3 dark:border-field-dark dark:bg-field-bg-dark"
				>
					<Text numberOfLines={1} className="flex-1 font-text text-16 text-fg dark:text-fg-dark">
						{labels.sort(sortLabel)}
					</Text>
					<Icon name="chevron-down" size={18} className="text-fg dark:text-fg-dark" />
				</Pressable>
			</View>
			{filtersOpen ? (
				<View className="mt-6 gap-4 bg-surface p-4 dark:bg-surface-dark">
					{props.filterPanel}
					<View className="flex-row gap-3">
						<View className="flex-1">
							<Btn label={labels.apply} onPress={props.onApply} />
						</View>
						<View className="flex-1">
							<Btn variant="secondary" label={labels.reset} onPress={props.onReset} />
						</View>
					</View>
				</View>
			) : rows.length === 0 ? (
				<View accessibilityLiveRegion="polite" className="items-start gap-4 py-8">
					<Text className="font-text text-16 text-fg dark:text-fg-dark">{labels.empty}</Text>
					<Btn variant="secondary" label={labels.reset} onPress={props.onReset} />
				</View>
			) : (
				<>
					{total === null ? null : (
						<Text accessibilityLiveRegion="polite" className={`${MUTED} pt-4 tabular-nums`}>
							{labels.found(total)}
						</Text>
					)}
					<Pager page={page} pages={pages} labels={labels} onPage={props.onPage} />
					{rows.map((row) => (
						<Row
							key={row.id}
							row={row}
							labels={labels}
							locale={props.locale}
							onOpen={props.onOpen}
						/>
					))}
					<Pager page={page} pages={pages} labels={labels} onPage={props.onPage} />
				</>
			)}
		</ScrollView>
	);
}
