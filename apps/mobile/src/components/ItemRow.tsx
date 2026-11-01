import { Image } from "expo-image";
import { memo, type ReactElement } from "react";
import { Pressable, Text, View } from "react-native";
import type { ItemJSON } from "../../../../packages/app-client/src/api";
import { type AppLocale, t } from "../i18n";
import { scoreBand } from "../theme/scoreBand";

/** Fixed height so lists can use getItemLayout (SDD 10 §7.1). */
export const ITEM_ROW_HEIGHT = 80;

// SDD 10 §7.3: the 96×144 thumb shown at 44 × 64 dp; not on the spacing scale, so a style.
const COVER = { width: 44, height: 64 } as const;

const BAND_CLASS = {
	"score-low": "bg-score-low",
	"score-mid": "bg-score-mid",
	"score-high": "bg-score-high",
	"score-none": "bg-score-none",
} as const;

export function formatScore(score: number | null | undefined, locale: AppLocale): string {
	if (score === null || score === undefined) return "–";
	return new Intl.NumberFormat(locale, {
		minimumFractionDigits: 1,
		maximumFractionDigits: 1,
	}).format(score);
}

export type ItemRowProps = {
	item: ItemJSON;
	locale: AppLocale;
	basis?: "trend" | "prior";
	/** Absent until /app/v1 items carry cover URLs; a reserved box keeps the row height. */
	coverThumbUrl?: string;
	onPress?: (item: ItemJSON) => void;
};

function ItemRowView({ item, locale, basis, coverThumbUrl, onPress }: ItemRowProps): ReactElement {
	const band = scoreBand(item.score ?? null);
	const score = formatScore(item.score, locale);
	return (
		<Pressable
			accessibilityRole="button"
			className="h-20 flex-row items-center gap-3 px-4"
			onPress={() => onPress?.(item)}
		>
			{coverThumbUrl ? (
				<Image
					testID="cover-thumb"
					source={{ uri: coverThumbUrl }}
					cachePolicy="memory-disk"
					style={COVER}
					accessibilityLabel={t(locale, "item.coverAlt", { title: item.title ?? item.slug })}
				/>
			) : (
				<View className="rounded-sm bg-surface dark:bg-surface-dark" style={COVER} />
			)}
			<View className="flex-1 gap-1">
				<Text className="text-12 text-muted dark:text-muted-dark">
					{t(locale, `catalog.title.${item.type}`)}
				</Text>
				<Text numberOfLines={2} className="text-16 text-fg dark:text-fg-dark">
					{item.title ?? item.slug}
				</Text>
				{basis === "prior" ? (
					<Text className="text-12 text-muted dark:text-muted-dark">
						{t(locale, "catalog.trending.popular")}
					</Text>
				) : null}
			</View>
			<View className="flex-row items-center gap-2">
				<View
					testID={`score-band-${band}`}
					className={`h-3 w-3 rounded-full ${BAND_CLASS[band]}`}
				/>
				<Text className="text-16 text-fg dark:text-fg-dark">{score}</Text>
			</View>
		</Pressable>
	);
}

export const ItemRow = memo(ItemRowView);
