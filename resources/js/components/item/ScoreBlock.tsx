import { Link, usePage } from "@inertiajs/react";

import { localized } from "../../lib/route";
import { useT } from "../../lib/useT";
import type { SharedProps } from "../../types";
import ScoreBadge from "../ui/ScoreBadge";

export function verdict(score: number | null): string | null {
	if (score === null) return null;
	if (score < 5) return "low";
	if (score < 7) return "good";
	if (score < 8) return "veryGood";
	if (score < 9) return "excellent";
	return "masterpiece";
}

// The final score with its band in words and the one-line explanation of where it comes from.
export default function ScoreBlock({
	score,
	ratingCount,
	unreleased = false,
}: {
	score: number | null;
	ratingCount: number;
	unreleased?: boolean;
}) {
	const t = useT();
	const { locale } = usePage<SharedProps>().props;
	const category = verdict(score);

	return (
		<div
			data-testid="score-block"
			className="flex items-center gap-4 border border-border bg-surface p-4"
		>
			<ScoreBadge score={unreleased ? null : score} size="lg" />
			<div className="grid gap-1">
				<p className="text-lg font-semibold">
					{unreleased
						? t("item.unreleased")
						: category === null
							? t("item.notEnough")
							: t(`item.verdict.${category}`)}
				</p>
				<p className="text-sm text-muted">
					{unreleased
						? t("item.unreleasedHint")
						: category === null
							? t("item.notEnoughHint", { min: 5 })
							: t("item.scoreFrom", {
									count: ratingCount,
									n: new Intl.NumberFormat(locale).format(ratingCount),
								})}
				</p>
				<Link
					href={localized("/how-score-works", locale)}
					className="w-fit text-sm font-medium text-primary underline-offset-2 hover:underline"
				>
					{t("item.howScoreWorks")} <span aria-hidden="true">→</span>
				</Link>
			</div>
		</div>
	);
}
