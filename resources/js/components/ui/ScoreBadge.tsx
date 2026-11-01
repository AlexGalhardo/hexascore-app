import { usePage } from "@inertiajs/react";

import { useT } from "../../lib/useT";
import type { SharedProps } from "../../types";

export type ScoreBand = "high" | "mid" | "low" | "none";

/** Score bands: below 5.0 low, 5.0 to 7.4 mid, 7.5 and above high, none without a score. */
export function scoreBand(score: number | null): ScoreBand {
	if (score === null) return "none";
	return score >= 7.5 ? "high" : score >= 5 ? "mid" : "low";
}

/** One decimal in the active locale, or "–" without a score (fewer than 5 ratings). */
export function formatScore(score: number | null, locale: string): string {
	if (score === null) return "–";
	return new Intl.NumberFormat(locale, {
		minimumFractionDigits: 1,
		maximumFractionDigits: 1,
	}).format(score);
}

const FILL: Record<ScoreBand, string> = {
	high: "bg-score-high",
	mid: "bg-score-mid",
	low: "bg-score-low",
	none: "bg-score-none",
};
const SIZE = {
	xs: "size-8 text-sm",
	sm: "size-10 text-base",
	md: "size-14 text-[22px]",
	lg: "size-24 text-[40px]",
	// Fills its box (the item tile footer): the parent sets the height, the number stays centred.
	fill: "size-full text-2xl",
} as const;

// The number is always printed on the band colour (ink on every fill, both themes).
export default function ScoreBadge({
	score,
	size = "sm",
}: {
	score: number | null;
	size?: keyof typeof SIZE;
}) {
	const t = useT();
	const { locale } = usePage<SharedProps>().props;
	return (
		<span
			role="img"
			aria-label={
				score === null
					? t("score.none")
					: t("score.outOf", { score: formatScore(score, locale) })
			}
			className={`inline-flex shrink-0 items-center justify-center rounded-xs font-bold text-on-score tabular-nums ${FILL[scoreBand(score)]} ${SIZE[size]}`}
		>
			{formatScore(score, locale)}
		</span>
	);
}
