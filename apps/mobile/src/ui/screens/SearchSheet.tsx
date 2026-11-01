import type { ReactElement } from "react";
import { Pressable, ScrollView, Text, TextInput, View } from "react-native";
import { Cover } from "../Cover";
import { Icon } from "../Icon";
import { ScoreBadge } from "../ScoreBadge";

/** Round 2 item 5: first 8 matches, "see all" goes to the full search page. */
export const MAX_ROWS = 8;

export interface SearchRow {
	id: string;
	title: string;
	year: number | null;
	typeLabel: string;
	coverUri: string | null;
	score: number | null;
}

export type SearchState =
	| { kind: "empty"; recent: string[]; trending: SearchRow[] }
	| { kind: "loading" }
	| { kind: "results"; rows: SearchRow[]; total: number };

export interface SearchLabels {
	placeholder: string;
	cancel: string;
	recent: string;
	removeRecent: (query: string) => string;
	trending: string;
	loading: string;
	noResults: (query: string) => string;
	firstOf: (shown: number, total: number) => string;
	seeAll: (query: string) => string;
	score: (row: SearchRow) => string;
}

export interface SearchSheetProps {
	query: string;
	onQueryChange: (query: string) => void;
	onCancel: () => void;
	onOpen: (row: SearchRow) => void;
	onRecent: (query: string) => void;
	onRemoveRecent: (query: string) => void;
	onSeeAll: () => void;
	state: SearchState;
	labels: SearchLabels;
	locale: string;
}

const SECTION =
	"px-4 pb-1 pt-4 font-text text-12 font-bold uppercase tracking-wider text-muted dark:text-muted-dark";
const MUTED = "font-text text-14 text-muted dark:text-muted-dark";

function Row({
	row,
	labels,
	locale,
	onOpen,
}: {
	row: SearchRow;
	labels: SearchLabels;
	locale: string;
	onOpen: (row: SearchRow) => void;
}): ReactElement {
	const scoreLabel = labels.score(row);
	return (
		<Pressable
			accessibilityRole="button"
			accessibilityLabel={[row.title, row.year, row.typeLabel, scoreLabel]
				.filter(Boolean)
				.join(", ")}
			onPress={() => onOpen(row)}
			className="min-h-11 flex-row items-center gap-3 px-4 py-2 active:bg-surface dark:active:bg-surface-dark"
		>
			<Cover uri={row.coverUri} title={row.title} className="w-8" />
			<View className="flex-1">
				<Text className="font-text text-16 font-semibold text-fg dark:text-fg-dark">
					{row.title}
				</Text>
				<Text className={MUTED}>{row.typeLabel}</Text>
			</View>
			{row.year !== null && <Text className={`${MUTED} tabular-nums`}>{row.year}</Text>}
			<ScoreBadge score={row.score} locale={locale} label={scoreLabel} />
		</Pressable>
	);
}

// Round 2 item 16, mobile: search opens as a full-screen sheet with the input focused.
export function SearchSheet({
	query,
	onQueryChange,
	onCancel,
	onOpen,
	onRecent,
	onRemoveRecent,
	onSeeAll,
	state,
	labels,
	locale,
}: SearchSheetProps): ReactElement {
	const rows = (list: SearchRow[]): ReactElement[] =>
		list
			.slice(0, MAX_ROWS)
			.map((row) => <Row key={row.id} row={row} labels={labels} locale={locale} onOpen={onOpen} />);
	return (
		<View className="flex-1 bg-bg dark:bg-bg-dark">
			<View className="flex-row items-center gap-2 border-b border-border px-4 dark:border-border-dark">
				<Icon name="search-outline" size={20} className="text-fg dark:text-fg-dark" />
				<TextInput
					autoFocus
					accessibilityLabel={labels.placeholder}
					value={query}
					onChangeText={onQueryChange}
					placeholder={labels.placeholder}
					returnKeyType="search"
					onSubmitEditing={onSeeAll}
					className="h-12 flex-1 font-text text-16 text-fg dark:text-fg-dark"
				/>
				<Pressable
					accessibilityRole="button"
					accessibilityLabel={labels.cancel}
					onPress={onCancel}
					className="h-11 justify-center"
				>
					<Text className="font-text text-16 font-semibold text-primary dark:text-primary-dark">
						{labels.cancel}
					</Text>
				</Pressable>
			</View>
			<ScrollView keyboardShouldPersistTaps="handled">
				{state.kind === "empty" && (
					<>
						{state.recent.length > 0 && <Text className={SECTION}>{labels.recent}</Text>}
						{state.recent.map((recent) => (
							<View key={recent} className="flex-row items-center px-4">
								<Pressable
									accessibilityRole="button"
									accessibilityLabel={recent}
									onPress={() => onRecent(recent)}
									className="min-h-11 flex-1 flex-row items-center gap-3"
								>
									<Icon name="time-outline" size={18} className="text-muted dark:text-muted-dark" />
									<Text className="font-text text-16 text-fg dark:text-fg-dark">{recent}</Text>
								</Pressable>
								<Pressable
									accessibilityRole="button"
									accessibilityLabel={labels.removeRecent(recent)}
									onPress={() => onRemoveRecent(recent)}
									className="h-11 w-11 items-center justify-center"
								>
									<Icon name="close" size={18} className="text-muted dark:text-muted-dark" />
								</Pressable>
							</View>
						))}
						{state.trending.length > 0 && <Text className={SECTION}>{labels.trending}</Text>}
						{rows(state.trending)}
					</>
				)}
				{state.kind === "loading" && (
					<Text accessibilityLiveRegion="polite" className={`${MUTED} px-4 py-4`}>
						{labels.loading}
					</Text>
				)}
				{state.kind === "results" && state.rows.length === 0 && (
					<Text accessibilityLiveRegion="polite" className={`${MUTED} px-4 py-4`}>
						{labels.noResults(query)}
					</Text>
				)}
				{state.kind === "results" && state.rows.length > 0 && (
					<>
						<Text accessibilityLiveRegion="polite" className={SECTION}>
							{labels.firstOf(Math.min(MAX_ROWS, state.rows.length), state.total)}
						</Text>
						{rows(state.rows)}
						<Pressable
							accessibilityRole="button"
							accessibilityLabel={labels.seeAll(query)}
							onPress={onSeeAll}
							className="min-h-11 flex-row items-center justify-between border-t border-border px-4 py-3 dark:border-border-dark"
						>
							<Text className="font-text text-16 font-semibold text-primary dark:text-primary-dark">
								{labels.seeAll(query)}
							</Text>
							<Icon
								name="arrow-forward"
								size={18}
								className="text-primary dark:text-primary-dark"
							/>
						</Pressable>
					</>
				)}
			</ScrollView>
		</View>
	);
}
