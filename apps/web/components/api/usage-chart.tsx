import type { UsageDay } from "@/lib/api/dashboard";

const W = 600;
const H = 120;
const GAP = 2;

/** Plain SVG bars; each bar has a <title> tooltip and the whole chart a text summary for screen readers. */
export function UsageChart({
	series,
	label,
	barLabel,
}: {
	series: UsageDay[];
	label: string;
	barLabel: (day: UsageDay) => string;
}) {
	const max = Math.max(1, ...series.map((d) => d.count));
	const bar = W / series.length;
	return (
		<svg
			viewBox={`0 0 ${W} ${H}`}
			role="img"
			aria-label={label}
			className="h-32 w-full text-fg"
			preserveAspectRatio="none"
		>
			<line x1="0" y1={H - 0.5} x2={W} y2={H - 0.5} stroke="var(--border)" />
			{series.map((d, i) => {
				const h = d.count === 0 ? 0 : Math.max(2, (d.count / max) * (H - 4));
				return (
					<rect
						key={d.day}
						x={i * bar + GAP / 2}
						y={H - h}
						width={bar - GAP}
						height={h}
						rx="1"
						className={i === series.length - 1 ? "fill-score-high" : "fill-current opacity-60"}
					>
						<title>{barLabel(d)}</title>
					</rect>
				);
			})}
		</svg>
	);
}
