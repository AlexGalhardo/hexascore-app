import type { ReactElement } from "react";
import { Pressable, ScrollView, Text, View } from "react-native";
import { Btn } from "../Btn";
import { Cover } from "../Cover";
import { Icon } from "../Icon";
import { ScoreBadge } from "../ScoreBadge";

export interface LibraryRow {
	id: string;
	title: string;
	typeLabel: string;
	year: number | null;
	coverUri: string | null;
	/** Community score. */
	score: number | null;
	/** The member's own overall rating, null when unrated. */
	myScore: number | null;
}

/** A status shelf or a user list: same shape, the wiring decides which. */
export interface LibraryGroup {
	key: string;
	title: string;
	total: number;
	/** First rows only; `total > rows.length` shows the see-all link. */
	rows: LibraryRow[];
}

export interface LibraryLabels {
	summary: (titles: number, rated: number) => string;
	/** Visible caption above the status picker. */
	statusField: string;
	status: (current: string) => string;
	filters: string;
	groupCount: (group: LibraryGroup) => string;
	seeAll: (group: LibraryGroup) => string;
	meta: (row: LibraryRow) => string;
	score: (row: LibraryRow) => string;
	empty: string;
	browse: string;
}

export interface LibraryPageProps {
	kicker: string;
	title: string;
	titles: number;
	rated: number;
	statusLabel: string;
	groups: LibraryGroup[];
	labels: LibraryLabels;
	locale: string;
	onStatus: () => void;
	onFilters: () => void;
	onOpen: (row: LibraryRow) => void;
	onSeeAll: (group: LibraryGroup) => void;
	onBrowse: () => void;
}

const MUTED = "font-text text-14 text-muted dark:text-muted-dark";
const FG = "text-fg dark:text-fg-dark";

function Row({
	row,
	labels,
	locale,
	onOpen,
}: {
	row: LibraryRow;
	labels: LibraryLabels;
	locale: string;
	onOpen: (row: LibraryRow) => void;
}): ReactElement {
	return (
		<Pressable
			accessibilityRole="button"
			accessibilityLabel={[row.title, labels.meta(row), labels.score(row)].join(", ")}
			onPress={() => onOpen(row)}
			className="min-h-11 flex-row items-center gap-4 border-b border-border py-3 active:bg-surface dark:border-border-dark dark:active:bg-surface-dark"
		>
			<Cover uri={row.coverUri} title={row.title} className="w-10" />
			<View className="flex-1">
				<Text className={`font-display text-18 font-bold ${FG}`}>{row.title}</Text>
				<Text className={MUTED}>{labels.meta(row)}</Text>
			</View>
			<ScoreBadge score={row.score} locale={locale} label={labels.score(row)} />
		</Pressable>
	);
}

function Group({ group, props }: { group: LibraryGroup; props: LibraryPageProps }): ReactElement {
	const { labels } = props;
	return (
		<View className="border-t-2 border-fg pt-4 dark:border-fg-dark">
			<View className="flex-row flex-wrap items-baseline gap-x-3">
				<Text accessibilityRole="header" className={`font-display text-24 font-bold ${FG}`}>
					{group.title}
				</Text>
				<Text className={`${MUTED} tabular-nums`}>{labels.groupCount(group)}</Text>
			</View>
			{group.rows.map((row) => (
				<Row key={row.id} row={row} labels={labels} locale={props.locale} onOpen={props.onOpen} />
			))}
			{group.total > group.rows.length && (
				<Pressable
					accessibilityRole="link"
					accessibilityLabel={labels.seeAll(group)}
					onPress={() => props.onSeeAll(group)}
					className="h-11 flex-row items-center gap-1 self-start"
				>
					<Text className="font-text text-14 font-semibold text-primary dark:text-primary-dark">
						{labels.seeAll(group)}
					</Text>
					<Icon name="arrow-forward" size={16} className="text-primary dark:text-primary-dark" />
				</Pressable>
			)}
		</View>
	);
}

// Round 2 library, mobile: status shelves stacked, each capped with a see-all link.
export function LibraryPage(props: LibraryPageProps): ReactElement {
	const { labels } = props;
	return (
		<ScrollView
			className="flex-1 bg-bg dark:bg-bg-dark"
			contentContainerClassName="gap-6 px-4 pb-8 pt-6"
		>
			<View className="gap-2">
				<Text className="font-text text-12 font-bold uppercase tracking-wider text-primary dark:text-primary-dark">
					{props.kicker}
				</Text>
				<Text accessibilityRole="header" className={`font-display text-36 font-bold ${FG}`}>
					{props.title}
				</Text>
				<Text className={`${MUTED} tabular-nums`}>{labels.summary(props.titles, props.rated)}</Text>
			</View>
			<Text className={`-mb-4 font-text text-14 font-semibold ${FG}`}>{labels.statusField}</Text>
			<View className="flex-row gap-2">
				<Pressable
					accessibilityRole="button"
					accessibilityLabel={labels.status(props.statusLabel)}
					onPress={props.onStatus}
					className="min-h-11 flex-1 flex-row items-center justify-between rounded-sm border border-field bg-field-bg px-3 dark:border-field-dark dark:bg-field-bg-dark"
				>
					<Text numberOfLines={1} className={`flex-1 font-text text-16 ${FG}`}>
						{props.statusLabel}
					</Text>
					<Icon name="chevron-down" size={18} className={FG} />
				</Pressable>
				<Btn variant="secondary" label={labels.filters} onPress={props.onFilters} />
			</View>
			{props.groups.length === 0 ? (
				<View accessibilityLiveRegion="polite" className="items-start gap-4 py-8">
					<Text className={`font-text text-16 ${FG}`}>{labels.empty}</Text>
					<Btn variant="secondary" label={labels.browse} onPress={props.onBrowse} />
				</View>
			) : (
				props.groups.map((group) => <Group key={group.key} group={group} props={props} />)
			)}
		</ScrollView>
	);
}
