const R = 100;
const RINGS = [25, 50, 75, 100];

/** Vertex i of a hexagon of the given radius, first axis pointing up, clockwise. */
export function hexVertex(i: number, radius: number): [number, number] {
	const angle = ((-90 + i * 60) * Math.PI) / 180;
	return [Math.round(Math.cos(angle) * radius * 100) / 100, Math.round(Math.sin(angle) * radius * 100) / 100];
}

export function radarPoints(values: (number | null)[]): string {
	return values.map((v, i) => hexVertex(i, (Math.max(0, Math.min(100, v ?? 0)) / 100) * R).join(",")).join(" ");
}

type Props = {
	values: (number | null)[];
	labels: string[];
	/** Text alternative read by screen readers (the SVG itself is role="img"). */
	summary: string;
	/** Tailwind text-* class; the polygon uses currentColor. */
	accentClass?: string;
	className?: string;
};

export function HexRadar({ values, labels, summary, accentClass = "text-fg", className = "" }: Props) {
	const hasData = values.some((v) => v !== null);
	return (
		<svg
			role="img"
			aria-label={summary}
			viewBox="-250 -130 500 260"
			className={`h-auto w-full max-w-xl ${accentClass} ${className}`}
		>
			<title>{summary}</title>
			<g className="text-border" stroke="currentColor" fill="none" strokeWidth="1">
				{RINGS.map((ring) => (
					<polygon key={ring} points={radarPoints(Array(6).fill(ring))} />
				))}
				{labels.map((label, i) => (
					<line key={label} x1="0" y1="0" x2={hexVertex(i, R)[0]} y2={hexVertex(i, R)[1]} />
				))}
			</g>
			{hasData ? (
				<polygon
					points={radarPoints(values)}
					fill="currentColor"
					fillOpacity="0.2"
					stroke="currentColor"
					strokeWidth="2"
					strokeLinejoin="round"
				/>
			) : null}
			{labels.map((label, i) => {
				const [x, y] = hexVertex(i, R + 14);
				const anchor = Math.abs(x) < 1 ? "middle" : x > 0 ? "start" : "end";
				return (
					<text
						key={label}
						x={x}
						y={y + (y < -R ? -4 : y > R ? 12 : 4)}
						textAnchor={anchor}
						className="fill-fg font-sans"
						fontSize="14"
					>
						{label}{" "}
						<tspan className="fill-muted font-mono font-bold" fontSize="14">
							{values[i] ?? "–"}
						</tspan>
					</text>
				);
			})}
		</svg>
	);
}
