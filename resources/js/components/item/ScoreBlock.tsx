import { Link, usePage } from "@inertiajs/react";

import { localized } from "../../lib/route";
import { useT } from "../../lib/useT";
import type { SharedProps } from "../../types";
import ScoreBadge, { scoreBand } from "../ui/ScoreBadge";

// The final score with its band in words and the one-line explanation of where it comes from.
export default function ScoreBlock({
	score,
	ratingCount,
}: {
	score: number | null;
	ratingCount: number;
}) {
	const t = useT();
	const { locale } = usePage<SharedProps>().props;
	const band = scoreBand(score);

	return (
		<div
			data-testid="score-block"
			className="flex items-center gap-4 border border-border bg-surface p-4"
		>
			<ScoreBadge score={score} size="lg" />
			<div className="grid gap-1">
				<p className="text-lg font-semibold">
					{band === "none" ? t("item.notEnough") : t(`item.band.${band}`)}
				</p>
				<p className="text-sm text-muted">
					{band === "none"
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
