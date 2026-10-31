import { type KeyboardEvent, type PointerEvent, useState } from "react";

import { useT } from "../../lib/useT";

const MAX = 10;
const STEP = 0.5;

const clamp = (v: number): number => Math.min(MAX, Math.max(STEP, v));

const STAR_PATH = "M12 2.5l2.9 6.3 6.9.8-5.1 4.7 1.4 6.8L12 17.6 5.9 21.1l1.4-6.8L2.2 9.6l6.9-.8z";

type StarFill = "full" | "half" | "empty";

// Inline SVG, so half stars render the same on every platform (the text glyph for a half star
// does not). The amber layer is clipped to half of the box for "half".
function Star({ fill }: { fill: StarFill }) {
	const svg = (cls: string) => (
		<svg viewBox="0 0 24 24" className={`size-6 shrink-0 ${cls}`} aria-hidden="true">
			<path d={STAR_PATH} fill="currentColor" />
		</svg>
	);
	return (
		<span aria-hidden="true" data-fill={fill} className="relative inline-flex size-6 shrink-0">
			{svg("text-gray-300")}
			{fill !== "empty" && (
				<span
					className={`absolute inset-y-0 left-0 overflow-hidden ${fill === "half" ? "w-1/2" : "w-full"}`}
				>
					{svg("text-amber-500")}
				</span>
			)}
		</span>
	);
}

export interface StarRatingProps {
	label: string;
	value: number | null;
	onChange: (value: number | null) => void;
	disabled?: boolean;
}

// Half-star slider (SDD 09 §7.4): 0.5-10 in 0.5 steps, null = unset.
export default function StarRating({ label, value, onChange, disabled = false }: StarRatingProps) {
	const t = useT();
	const [hover, setHover] = useState<number | null>(null);
	const shown = hover ?? value ?? 0;

	const set = (next: number | null) => {
		if (!disabled) onChange(next);
	};

	const fromPointer = (e: PointerEvent<HTMLElement>): number => {
		const box = e.currentTarget.getBoundingClientRect();
		const ratio = box.width > 0 ? (e.clientX - box.left) / box.width : 0;
		return clamp(Math.ceil(ratio * MAX * 2) / 2);
	};

	const onKeyDown = (e: KeyboardEvent<HTMLElement>) => {
		const current = value ?? 0;
		const moves: Record<string, number | null> = {
			ArrowLeft: current - STEP,
			ArrowDown: current - STEP,
			ArrowRight: current + STEP,
			ArrowUp: current + STEP,
			PageDown: current - 1,
			PageUp: current + 1,
			Home: STEP,
			End: MAX,
			Delete: null,
		};
		if (!(e.key in moves)) return;
		e.preventDefault();
		const next = moves[e.key];
		// Decrementing at the minimum stays at the minimum rather than clearing.
		set(next === null ? null : clamp(next));
	};

	const valueText = value === null ? t("rating.unset") : t("rating.valuetext", { value });

	return (
		<div className="flex flex-wrap items-center gap-2">
			<span className="w-full text-sm font-medium sm:w-28 sm:shrink-0">{label}</span>
			<button
				type="button"
				className="size-11 rounded border"
				aria-label={`${label} -`}
				disabled={disabled}
				tabIndex={-1}
				onClick={() => set(clamp((value ?? 0) - STEP))}
			>
				-
			</button>
			<div
				role="slider"
				tabIndex={disabled ? -1 : 0}
				aria-label={label}
				aria-valuemin={0}
				aria-valuemax={MAX}
				aria-valuenow={value ?? 0}
				aria-valuetext={valueText}
				aria-disabled={disabled}
				className="flex min-h-11 flex-1 cursor-pointer items-center text-xl outline-offset-2 focus-visible:outline-2"
				onKeyDown={onKeyDown}
				onPointerMove={(e) => !disabled && setHover(fromPointer(e))}
				onPointerLeave={() => setHover(null)}
				onClick={(e) => set(fromPointer(e as unknown as PointerEvent<HTMLElement>))}
			>
				{Array.from({ length: MAX }, (_, i) => (
					<Star
						// biome-ignore lint/suspicious/noArrayIndexKey: fixed list of 10 positions
						key={i}
						fill={shown >= i + 1 ? "full" : shown >= i + 0.5 ? "half" : "empty"}
					/>
				))}
			</div>
			<button
				type="button"
				className="size-11 rounded border"
				aria-label={`${label} +`}
				disabled={disabled}
				tabIndex={-1}
				onClick={() => set(clamp((value ?? 0) + STEP))}
			>
				+
			</button>
			<output className="w-10 text-right tabular-nums">
				{value === null ? <span className="text-muted">–</span> : value.toFixed(1)}
			</output>
		</div>
	);
}
