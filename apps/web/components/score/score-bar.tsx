import { scoreBand } from "@/lib/categories";
import { BAND_BG } from "./score-badge";

type Props = { label: string; value: number | null };

export function ScoreBar({ label, value }: Props) {
	return (
		<div className="grid grid-cols-[minmax(0,1fr)_2.5rem] items-center gap-x-3 gap-y-1">
			<span className="truncate text-sm">{label}</span>
			<span className="text-right font-mono text-sm font-semibold tabular-nums">{value ?? "–"}</span>
			<span aria-hidden="true" className="col-span-2 h-1.5 overflow-hidden rounded-full bg-surface">
				<span
					className={`block h-full rounded-full ${BAND_BG[scoreBand(value)]}`}
					style={{ width: `${value ?? 0}%` }}
				/>
			</span>
		</div>
	);
}
