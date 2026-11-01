import { usePage } from "@inertiajs/react";

import { useT } from "../../lib/useT";
import type { SharedProps } from "../../types";
import { formatScore, scoreBand } from "../ui/ScoreBadge";

const BAR_FILL = {
	high: "fill-score-high",
	mid: "fill-score-mid",
	low: "fill-score-low",
	none: "fill-score-none",
} as const;

const C = 100;
const R = 80;

// Point of axis i (of n, clockwise from the top) at value v on the 0..10 scale.
function point(i: number, n: number, v: number): string {
	const a = -Math.PI / 2 + (2 * Math.PI * i) / n;
	const r = (R * v) / 10;
	return `${(C + r * Math.cos(a)).toFixed(1)},${(C + r * Math.sin(a)).toFixed(1)}`;
}

const ring = (n: number, v: number): string =>
	Array.from({ length: n }, (_, i) => point(i, n, v)).join(" ");

interface Props {
	keys: string[];
	labels: string[];
	averages: Record<string, number>;
}

// Radar plus bars with the number printed on every row; SVG attributes only (no inline style, CSP).
export default function CategoryAverages({ keys, labels, averages }: Props) {
	const t = useT();
	const { locale } = usePage<SharedProps>().props;
	const has = keys.some((k) => averages[k] !== undefined);
	const summary = keys
		.map((k, i) => `${labels[i]} ${formatScore(averages[k] ?? null, locale)}`)
		.join(", ");

	return (
		<div className="grid items-center gap-6 sm:grid-cols-[12rem_minmax(0,1fr)] sm:gap-10">
			<svg
				viewBox="0 0 200 200"
				role="img"
				aria-label={has ? t("score.radar", { summary }) : t("score.radarEmpty")}
				className="mx-auto size-48"
			>
				<polygon points={ring(keys.length, 10)} className="fill-none stroke-border" />
				<polygon points={ring(keys.length, 5)} className="fill-none stroke-border" />
				{has && (
					<polygon
						data-testid="radar-shape"
						points={keys
							.map((k, i) => point(i, keys.length, averages[k] ?? 0))
							.join(" ")}
						className="fill-primary/20 stroke-primary stroke-2"
					/>
				)}
			</svg>
			<dl className="grid gap-3">
				{keys.map((k, i) => {
					const v = averages[k] ?? null;
					return (
						<div
							key={k}
							className="grid grid-cols-[7rem_minmax(0,1fr)_2.5rem] items-center gap-3 text-sm"
						>
							<dt>{labels[i]}</dt>
							<dd className="contents">
								<svg
									viewBox="0 0 100 8"
									preserveAspectRatio="none"
									aria-hidden="true"
									className="h-2 w-full"
								>
									<rect width="100" height="8" className="fill-border" />
									{v !== null && (
										<rect
											width={v * 10}
											height="8"
											className={BAR_FILL[scoreBand(v)]}
										/>
									)}
								</svg>
								<span className="text-right font-semibold tabular-nums">
									{formatScore(v, locale)}
								</span>
							</dd>
						</div>
					);
				})}
			</dl>
		</div>
	);
}
