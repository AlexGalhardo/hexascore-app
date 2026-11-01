import type { ReactElement } from "react";
import { Pressable, Text, View } from "react-native";
import { Cover } from "../Cover";
import { Icon } from "../Icon";
import { ScoreBadge } from "../ScoreBadge";

/** Round 2 home: three trending cards, then a grid of six (two rows of three). */
export const TRENDING = 3;
export const GRID = 6;

export type ItemType = "game" | "movie" | "series" | "book";

export interface HomeItem {
	id: string;
	title: string;
	year: number | null;
	coverUri: string | null;
	score: number | null;
	ratingCount: number;
}

export interface HomeSection {
	type: ItemType;
	title: string;
	typeLabel: string;
	trending: HomeItem[];
	grid: HomeItem[];
}

export interface HomeLabels {
	seeAll: (section: HomeSection) => string;
	rank: (position: number, item: HomeItem) => string;
	/** "High score", "Not enough ratings", ... */
	verdict: (item: HomeItem) => string;
	/** "from 213 ratings", "3 of 5 ratings", ... */
	ratings: (item: HomeItem) => string;
	score: (item: HomeItem) => string;
}

export interface HomeSectionsProps {
	sections: HomeSection[];
	labels: HomeLabels;
	locale: string;
	onOpen: (item: HomeItem, type: ItemType) => void;
	onSeeAll: (type: ItemType) => void;
}

// Literal class names so NativeWind sees every accent at build time.
const ACCENT_BG: Record<ItemType, string> = {
	game: "bg-accent-game dark:bg-accent-game-dark",
	movie: "bg-accent-movie dark:bg-accent-movie-dark",
	series: "bg-accent-series dark:bg-accent-series-dark",
	book: "bg-accent-book dark:bg-accent-book-dark",
};
const ACCENT_BORDER: Record<ItemType, string> = {
	game: "border-t-accent-game dark:border-t-accent-game-dark",
	movie: "border-t-accent-movie dark:border-t-accent-movie-dark",
	series: "border-t-accent-series dark:border-t-accent-series-dark",
	book: "border-t-accent-book dark:border-t-accent-book-dark",
};

const MUTED = "font-text text-14 text-muted dark:text-muted-dark";

interface CardProps {
	item: HomeItem;
	section: HomeSection;
	position: number;
	labels: HomeLabels;
	locale: string;
	onOpen: (item: HomeItem, type: ItemType) => void;
}

function TrendingCard({
	item,
	section,
	position,
	labels,
	locale,
	onOpen,
}: CardProps): ReactElement {
	const meta = [item.year, section.typeLabel].filter((part) => part !== null).join(" · ");
	return (
		<Pressable
			accessibilityRole="button"
			accessibilityLabel={[labels.rank(position, item), item.title, meta, labels.score(item)].join(
				", ",
			)}
			onPress={() => onOpen(item, section.type)}
			className={`flex-row gap-3 border border-t-2 border-border bg-surface p-3 active:opacity-80 dark:border-border-dark dark:bg-surface-dark ${ACCENT_BORDER[section.type]}`}
		>
			<Cover uri={item.coverUri} title={item.title} className="w-16" />
			<View className="flex-1 gap-1">
				{/* fg, not the accent: accent-book on surface is 4.28:1 in light, below AA for 12 px text. */}
				<Text className="font-text text-12 font-bold text-fg dark:text-fg-dark">
					{labels.rank(position, item)}
				</Text>
				<Text className="font-display text-20 font-bold text-fg dark:text-fg-dark">
					{item.title}
				</Text>
				<Text className={MUTED}>{meta}</Text>
				<View className="mt-1 flex-row items-center gap-2">
					<ScoreBadge score={item.score} locale={locale} label={labels.score(item)} />
					<View className="flex-1">
						<Text className="font-text text-14 font-semibold text-fg dark:text-fg-dark">
							{labels.verdict(item)}
						</Text>
						<Text className={MUTED}>{labels.ratings(item)}</Text>
					</View>
				</View>
			</View>
		</Pressable>
	);
}

function GridCell({
	item,
	section,
	labels,
	locale,
	onOpen,
}: Omit<CardProps, "position">): ReactElement {
	return (
		<Pressable
			accessibilityRole="button"
			accessibilityLabel={[item.title, item.year, labels.score(item)].filter(Boolean).join(", ")}
			onPress={() => onOpen(item, section.type)}
			className="w-1/3 gap-1 px-1 pb-4 active:opacity-80"
		>
			<Cover uri={item.coverUri} title={item.title} className="w-full" />
			<Text numberOfLines={2} className="font-text text-14 text-fg dark:text-fg-dark">
				{item.title}
			</Text>
			<View className="flex-row items-center justify-between">
				<Text className={`${MUTED} tabular-nums`}>{item.year ?? ""}</Text>
				<ScoreBadge score={item.score} locale={locale} label={labels.score(item)} />
			</View>
		</Pressable>
	);
}

export function HomeSections({
	sections,
	labels,
	locale,
	onOpen,
	onSeeAll,
}: HomeSectionsProps): ReactElement {
	return (
		<View className="gap-8 py-4">
			{sections.map((section) => (
				<View key={section.type} testID={`home-section-${section.type}`} className="gap-3 px-4">
					<View className="flex-row items-center gap-2">
						<View className={`h-3 w-3 ${ACCENT_BG[section.type]}`} />
						<Text
							accessibilityRole="header"
							className="flex-1 font-display text-24 font-bold text-fg dark:text-fg-dark"
						>
							{section.title}
						</Text>
						<Pressable
							accessibilityRole="link"
							accessibilityLabel={labels.seeAll(section)}
							onPress={() => onSeeAll(section.type)}
							className="h-11 flex-row items-center gap-1"
						>
							<Text className="font-text text-14 font-semibold text-primary dark:text-primary-dark">
								{labels.seeAll(section)}
							</Text>
							<Icon
								name="arrow-forward"
								size={16}
								className="text-primary dark:text-primary-dark"
							/>
						</Pressable>
					</View>
					{section.trending.slice(0, TRENDING).map((item, index) => (
						<TrendingCard
							key={item.id}
							item={item}
							section={section}
							position={index + 1}
							labels={labels}
							locale={locale}
							onOpen={onOpen}
						/>
					))}
					{section.grid.length > 0 && (
						<View className="-mx-1 flex-row flex-wrap">
							{section.grid.slice(0, GRID).map((item) => (
								<GridCell
									key={item.id}
									item={item}
									section={section}
									labels={labels}
									locale={locale}
									onOpen={onOpen}
								/>
							))}
						</View>
					)}
				</View>
			))}
		</View>
	);
}
