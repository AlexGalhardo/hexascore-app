import { type ScoreBand, scoreBand } from "@/lib/categories";

// Dark text on every band: the band colours are mid-lightness, so black keeps ≥ 4.5:1 at any size.
export const BAND_BG: Record<ScoreBand, string> = {
	low: "bg-score-low",
	mid: "bg-score-mid",
	high: "bg-score-high",
	none: "bg-score-none",
};

const SIZES = {
	sm: "h-7 min-w-9 px-1 text-sm",
	md: "h-10 min-w-12 px-1.5 text-xl",
	lg: "h-20 min-w-20 px-2 text-5xl",
} as const;

type Props = {
	score: number | null;
	/** Accessible text, e.g. "Score 82 out of 100" or "Not enough ratings yet". */
	label: string;
	size?: keyof typeof SIZES;
	className?: string;
};

export function ScoreBadge({ score, label, size = "md", className = "" }: Props) {
	return (
		<span
			role="img"
			aria-label={label}
			title={label}
			className={`inline-flex shrink-0 items-center justify-center rounded-md font-mono font-bold tabular-nums text-black ${BAND_BG[scoreBand(score)]} ${SIZES[size]} ${className}`}
		>
			{score ?? "–"}
		</span>
	);
}
