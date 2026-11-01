import type { ReactElement, ReactNode } from "react";
import { Pressable, ScrollView, Text, View } from "react-native";
import { scoreBand } from "../../theme/scoreBand";
import { Btn } from "../Btn";
import { Cover } from "../Cover";
import { formatScore } from "../format";
import { Icon } from "../Icon";
import { RatingInput } from "../RatingInput";
import { ScoreBadge } from "../ScoreBadge";

export interface CategoryScore {
	key: string;
	label: string;
	score: number | null;
}

export interface ItemDetailView {
	title: string;
	/** "Game · 2029": type label and year, already localised. */
	kicker: string;
	byline: string | null;
	coverUri: string | null;
	score: number | null;
	ratingCount: number;
	genres: string[];
	synopsis: string | null;
	categories: CategoryScore[];
}

export interface SimilarItem {
	id: string;
	title: string;
	year: number | null;
	coverUri: string | null;
	score: number | null;
}

/** The viewer's own rating; null when logged out (Rate then routes to login). */
export interface RatingDraft {
	values: Record<string, number | null>;
	overall: number | null;
	saving: boolean;
}

export interface ItemDetailLabels {
	verdict: (score: number | null) => string;
	ratings: (count: number) => string;
	score: (score: number | null) => string;
	rate: string;
	addToList: string;
	categoryAverages: string;
	synopsis: string;
	noSynopsis: string;
	rateTitle: (title: string) => string;
	overall: string;
	save: string;
	moreLikeThis: string;
}

export interface ItemDetailProps {
	item: ItemDetailView;
	/** Current library status ("Want to play") or the "set status" prompt; the button is hidden without it. */
	statusLabel?: string;
	similar: SimilarItem[];
	rating: RatingDraft | null;
	/** The viewer's saved overall score, shown under the actions. */
	ownScore?: { label: string; score: number } | null;
	labels: ItemDetailLabels;
	locale: string;
	onRate: () => void;
	onStatus?: () => void;
	onAddToList?: () => void;
	onRateCategory: (key: string, value: number) => void;
	onSave: () => void;
	onOpenSimilar: (item: SimilarItem) => void;
	/** Screen-specific blocks (entities, editions, attribution) after the sections. */
	children?: ReactNode;
}

const BAR = {
	"score-low": "bg-score-low",
	"score-mid": "bg-score-mid",
	"score-high": "bg-score-high",
	"score-none": "bg-score-none",
} as const;

const H2 = "font-display text-24 font-bold text-fg dark:text-fg-dark";
const MUTED = "font-text text-14 text-muted dark:text-muted-dark";
const RULE = "gap-3 border-t border-fg pt-4 dark:border-fg-dark";

function CategoryBar({
	category,
	labels,
	locale,
}: {
	category: CategoryScore;
	labels: ItemDetailLabels;
	locale: string;
}): ReactElement {
	return (
		<View
			accessible
			accessibilityLabel={`${category.label}, ${labels.score(category.score)}`}
			className="min-h-8 flex-row items-center gap-3"
		>
			<Text className="w-28 font-text text-14 text-fg dark:text-fg-dark">{category.label}</Text>
			<View className="h-2 flex-1 bg-surface dark:bg-surface-dark">
				{category.score !== null && (
					// Data-driven width has no utility class; a native style object, not a web style attribute.
					<View
						className={`h-2 ${BAR[scoreBand(category.score)]}`}
						style={{ width: `${category.score * 10}%` }}
					/>
				)}
			</View>
			<Text className="w-10 text-right font-display text-16 font-bold tabular-nums text-fg dark:text-fg-dark">
				{formatScore(category.score, locale)}
			</Text>
		</View>
	);
}

// v1 item page (android__item__default + web-mobile__item__logged-in): rating panel only when logged in.
export function ItemDetail(props: ItemDetailProps): ReactElement {
	const { item, rating, labels, locale } = props;
	return (
		<ScrollView
			testID="item"
			className="flex-1 bg-bg dark:bg-bg-dark"
			contentContainerClassName="gap-6 px-4 pb-8 pt-4"
		>
			<View className="flex-row gap-4">
				<Cover uri={item.coverUri} title={item.title} className="w-28" />
				<View className="flex-1 justify-center gap-1">
					<Text className="font-text text-12 font-bold uppercase tracking-wider text-primary dark:text-primary-dark">
						{item.kicker}
					</Text>
					<Text
						accessibilityRole="header"
						className="font-display text-30 font-bold text-fg dark:text-fg-dark"
					>
						{item.title}
					</Text>
					{item.byline !== null && <Text className={MUTED}>{item.byline}</Text>}
					<View className="mt-2 flex-row items-center gap-3">
						<View testID={`score-band-${scoreBand(item.score)}`}>
							<ScoreBadge
								score={item.score}
								locale={locale}
								label={labels.score(item.score)}
								size="lg"
							/>
						</View>
						<Text className={`${MUTED} flex-1`}>
							{`${labels.verdict(item.score)} · ${labels.ratings(item.ratingCount)}`}
						</Text>
					</View>
				</View>
			</View>
			{item.genres.length > 0 && (
				<View className="flex-row flex-wrap gap-2">
					{item.genres.map((genre) => (
						<View
							key={genre}
							testID="genre-chip"
							className="min-h-11 justify-center rounded-full border border-fg px-4 dark:border-fg-dark"
						>
							<Text className="font-text text-14 text-fg dark:text-fg-dark">{genre}</Text>
						</View>
					))}
				</View>
			)}
			<View className="flex-row gap-2">
				<View className="flex-1">
					<Btn label={labels.rate} onPress={props.onRate} />
				</View>
				{props.statusLabel !== undefined && props.onStatus && (
					<View className="flex-1">
						<Btn variant="secondary" label={props.statusLabel} onPress={props.onStatus} />
					</View>
				)}
				{props.onAddToList && (
					<Pressable
						accessibilityRole="button"
						accessibilityLabel={labels.addToList}
						onPress={props.onAddToList}
						className="h-11 w-11 items-center justify-center rounded-sm border border-b-2 border-fg bg-field-bg active:border-b dark:border-fg-dark dark:bg-field-bg-dark"
					>
						<Icon name="add" size={22} className="text-fg dark:text-fg-dark" />
					</Pressable>
				)}
			</View>
			{props.ownScore && (
				<View className="flex-row items-center justify-between">
					<Text className="font-text text-16 font-semibold text-fg dark:text-fg-dark">
						{props.ownScore.label}
					</Text>
					<ScoreBadge
						score={props.ownScore.score}
						locale={locale}
						label={`${props.ownScore.label}, ${labels.score(props.ownScore.score)}`}
					/>
				</View>
			)}
			{rating !== null && (
				<View className="gap-4 border-t-2 border-primary bg-surface p-4 dark:border-primary-dark dark:bg-surface-dark">
					<Text className={H2}>{labels.rateTitle(item.title)}</Text>
					{item.categories.map((category) => (
						<View key={category.key} className="gap-1">
							<Text className="font-text text-14 font-semibold text-fg dark:text-fg-dark">
								{category.label}
							</Text>
							<RatingInput
								label={category.label}
								value={rating.values[category.key] ?? null}
								locale={locale}
								onChange={(value) => props.onRateCategory(category.key, value)}
							/>
						</View>
					))}
					<View className="flex-row items-center justify-between">
						<Text className="font-text text-16 font-semibold text-fg dark:text-fg-dark">
							{labels.overall}
						</Text>
						<ScoreBadge
							score={rating.overall}
							locale={locale}
							label={`${labels.overall}, ${labels.score(rating.overall)}`}
						/>
					</View>
					<Btn label={labels.save} disabled={rating.saving} onPress={props.onSave} />
				</View>
			)}
			{item.categories.length > 0 && (
				<View className={RULE}>
					<Text className={H2}>{labels.categoryAverages}</Text>
					{item.categories.map((category) => (
						<CategoryBar key={category.key} category={category} labels={labels} locale={locale} />
					))}
				</View>
			)}
			<View className={RULE}>
				<Text className={H2}>{labels.synopsis}</Text>
				{item.synopsis === null ? (
					<Text className="font-display text-16 italic text-muted dark:text-muted-dark">
						{labels.noSynopsis}
					</Text>
				) : (
					<Text className="font-text text-16 text-fg dark:text-fg-dark">{item.synopsis}</Text>
				)}
			</View>
			{props.similar.length > 0 && (
				<View className={RULE}>
					<Text className={H2}>{labels.moreLikeThis}</Text>
					<View className="flex-row gap-3">
						{props.similar.slice(0, 3).map((similar) => (
							<Pressable
								key={similar.id}
								accessibilityRole="button"
								accessibilityLabel={[similar.title, similar.year, labels.score(similar.score)]
									.filter(Boolean)
									.join(", ")}
								onPress={() => props.onOpenSimilar(similar)}
								className="w-1/3 flex-1 gap-1 active:opacity-80"
							>
								<Cover uri={similar.coverUri} title={similar.title} className="w-full" />
								<Text numberOfLines={2} className="font-text text-14 text-fg dark:text-fg-dark">
									{similar.title}
								</Text>
								<View className="flex-row items-center justify-between">
									<Text className={`${MUTED} tabular-nums`}>{similar.year ?? ""}</Text>
									<ScoreBadge
										score={similar.score}
										locale={locale}
										label={labels.score(similar.score)}
									/>
								</View>
							</Pressable>
						))}
					</View>
				</View>
			)}
			{props.children}
		</ScrollView>
	);
}
