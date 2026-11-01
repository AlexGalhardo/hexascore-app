import type { ReactElement, ReactNode } from "react";
import { Pressable, Text, View } from "react-native";
import { Cover } from "./Cover";
import { ScoreBadge } from "./ScoreBadge";

export interface ItemRowData {
	id: string;
	title: string;
	/** Already-formatted secondary line (type · year · role). */
	meta: string;
	coverUri: string | null;
	score: number | null;
	scoreLabel: string;
}

export interface ItemRowProps {
	row: ItemRowData;
	locale: string;
	onOpen: (row: ItemRowData) => void;
	/** Trailing controls (reorder, remove) next to the row, outside its press target. */
	trailing?: ReactNode;
}

export function ItemRow({ row, locale, onOpen, trailing }: ItemRowProps): ReactElement {
	return (
		<View className="flex-row items-center gap-2 border-b border-border dark:border-border-dark">
			<Pressable
				accessibilityRole="button"
				accessibilityLabel={[row.title, row.meta, row.scoreLabel].join(", ")}
				onPress={() => onOpen(row)}
				className="min-h-11 flex-1 flex-row items-center gap-4 py-3 active:bg-surface dark:active:bg-surface-dark"
			>
				<Cover uri={row.coverUri} title={row.title} className="w-10" />
				<View className="flex-1">
					<Text className="font-display text-18 font-bold text-fg dark:text-fg-dark">
						{row.title}
					</Text>
					<Text className="font-text text-14 text-muted dark:text-muted-dark">{row.meta}</Text>
				</View>
				<ScoreBadge score={row.score} locale={locale} label={row.scoreLabel} />
			</Pressable>
			{trailing}
		</View>
	);
}
