import { type KeyboardEvent, type PointerEvent, useState } from "react";

import { useT } from "../../lib/useT";

const MAX = 10;
const STEP = 0.5;

const clamp = (v: number): number => Math.min(MAX, Math.max(STEP, v));

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
		<div className="flex items-center gap-2">
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
				className="flex min-h-11 flex-1 cursor-pointer items-center outline-offset-2 focus-visible:outline-2"
				onKeyDown={onKeyDown}
				onPointerMove={(e) => !disabled && setHover(fromPointer(e))}
				onPointerLeave={() => setHover(null)}
				onClick={(e) => set(fromPointer(e as unknown as PointerEvent<HTMLElement>))}
			>
				{Array.from({ length: MAX }, (_, i) => (
					<span
						// biome-ignore lint/suspicious/noArrayIndexKey: fixed list of 10 positions
						key={i}
						aria-hidden="true"
						className={
							shown >= i + 1
								? "text-amber-500"
								: shown >= i + 0.5
									? "text-amber-300"
									: "text-gray-300"
						}
					>
						{shown >= i + 1 ? "★" : shown >= i + 0.5 ? "⯪" : "☆"}
					</span>
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
			<output className="w-10 tabular-nums">{value === null ? "-" : value.toFixed(1)}</output>
		</div>
	);
}
