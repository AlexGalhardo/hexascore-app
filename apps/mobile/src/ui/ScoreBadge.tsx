import type { ReactElement } from "react";
import { Text, View } from "react-native";
import { scoreBand } from "../theme/scoreBand";
import { formatScore } from "./format";

const FILL = {
	"score-low": "bg-score-low",
	"score-mid": "bg-score-mid",
	"score-high": "bg-score-high",
	"score-none": "bg-score-none",
} as const;

export interface ScoreBadgeProps {
	score: number | null;
	locale: string;
	label: string;
	size?: "md" | "lg";
}

// Ink number on every fill in both themes (ROUND-2 contrast table), so no dark: variant.
export function ScoreBadge({ score, locale, label, size = "md" }: ScoreBadgeProps): ReactElement {
	const box = size === "lg" ? "h-14 w-14" : "h-9 w-11";
	const text = size === "lg" ? "text-24" : "text-16";
	return (
		<View
			testID="score-badge"
			accessible
			accessibilityLabel={label}
			className={`${box} items-center justify-center rounded-sm ${FILL[scoreBand(score)]}`}
		>
			<Text className={`${text} font-bold tabular-nums text-on-score`}>
				{formatScore(score, locale)}
			</Text>
		</View>
	);
}
