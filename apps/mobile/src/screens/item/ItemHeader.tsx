import { Image } from "expo-image";
import type { ReactElement } from "react";
import { Text, View } from "react-native";
import type { ItemJSON } from "../../../../../packages/app-client/src/api";
import { formatScore } from "../../components/ItemRow";
import { type AppLocale, t } from "../../i18n";
import { isCatalogType, SINGULAR } from "../../navigation/catalog";
import { scoreBand } from "../../theme/scoreBand";

// The 600 px cover scaled to a 2:3 box (SDD 10 §7.3); not on the spacing scale.
const COVER = { width: 160, height: 240 } as const;

const BAND_CLASS = {
	"score-low": "bg-score-low",
	"score-mid": "bg-score-mid",
	"score-high": "bg-score-high",
	"score-none": "bg-score-none",
} as const;

export function ItemHeader({
	item,
	locale,
	coverUrl,
}: {
	item: ItemJSON;
	locale: AppLocale;
	/** Absent until /app/v1 items carry cover URLs (FIX-app-item-cover-urls); the box is reserved either way. */
	coverUrl?: string;
}): ReactElement {
	const band = scoreBand(item.score ?? null);
	const singular = isCatalogType(item.type) ? SINGULAR[item.type] : item.type;
	const level = item.confidence ? t(locale, `item.confidenceLevel.${item.confidence}`) : "–";
	const categories = Object.entries(item.scoreByCategory ?? {});
	return (
		<View className="gap-3 px-4 pt-4">
			{coverUrl ? (
				<Image
					source={{ uri: coverUrl }}
					cachePolicy="memory-disk"
					style={COVER}
					accessibilityLabel={t(locale, "item.coverAlt", { title: item.title ?? item.slug })}
				/>
			) : (
				<View className="rounded-md bg-surface dark:bg-surface-dark" style={COVER} />
			)}
			<Text accessibilityRole="header" className="text-24 text-fg dark:text-fg-dark">
				{item.title ?? item.slug}
			</Text>
			<View className="flex-row items-center gap-2">
				<View
					testID={`score-band-${band}`}
					className={`h-3 w-3 rounded-full ${BAND_CLASS[band]}`}
				/>
				<Text className="text-20 text-fg dark:text-fg-dark">{formatScore(item.score, locale)}</Text>
				<Text className="text-14 text-muted dark:text-muted-dark">
					{t(locale, "catalog.ratingCount", { count: item.ratingCount ?? 0 })}
				</Text>
			</View>
			<Text className="text-14 text-muted dark:text-muted-dark">
				{t(locale, "item.confidence", { level })}
			</Text>
			{categories.length > 0 ? (
				<View className="gap-1">
					<Text className="text-16 text-fg dark:text-fg-dark">
						{t(locale, "item.categoryScores")}
					</Text>
					{categories.map(([key, value]) => (
						<View key={key} className="flex-row justify-between">
							<Text className="text-14 text-fg dark:text-fg-dark">
								{t(locale, `categories.${singular}.${key}`)}
							</Text>
							<Text className="text-14 text-fg dark:text-fg-dark">
								{formatScore(value, locale)}
							</Text>
						</View>
					))}
				</View>
			) : null}
		</View>
	);
}
