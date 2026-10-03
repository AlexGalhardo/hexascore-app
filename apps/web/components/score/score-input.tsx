"use client";

import { type KeyboardEvent, useEffect, useId, useState } from "react";

type Props = {
	label: string;
	value: number;
	onChange: (value: number) => void;
	disabled?: boolean;
};

const clamp = (n: number) => Math.min(100, Math.max(0, Math.round(n)));

// Same keys on both controls (browsers differ on PageUp/PageDown for range and do nothing for number inputs).
const STEPS: Record<string, (v: number) => number> = {
	ArrowUp: (v) => v + 1,
	ArrowRight: (v) => v + 1,
	ArrowDown: (v) => v - 1,
	ArrowLeft: (v) => v - 1,
	PageUp: (v) => v + 10,
	PageDown: (v) => v - 10,
	Home: () => 0,
	End: () => 100,
};

/** Slider + number field kept in sync, 0–100 integers (design system "Score input"). */
export function ScoreInput({ label, value, onChange, disabled }: Props) {
	const id = useId();
	const [draft, setDraft] = useState(String(value));

	useEffect(() => setDraft(String(value)), [value]);

	function onKeyDown(event: KeyboardEvent<HTMLInputElement>): void {
		const step = STEPS[event.key];
		if (!step) return;
		event.preventDefault();
		onChange(clamp(step(value)));
	}

	return (
		<div className="grid grid-cols-[1fr_4.5rem] items-center gap-x-3 gap-y-1">
			<label htmlFor={`${id}-n`} className="col-span-2 text-sm font-medium">
				{label}
			</label>
			<input
				type="range"
				min={0}
				max={100}
				step={1}
				value={value}
				disabled={disabled}
				aria-label={label}
				onKeyDown={onKeyDown}
				onChange={(e) => onChange(clamp(Number(e.target.value)))}
				className="h-11 w-full cursor-pointer accent-current disabled:cursor-not-allowed"
			/>
			<input
				id={`${id}-n`}
				type="number"
				inputMode="numeric"
				min={0}
				max={100}
				step={1}
				value={draft}
				disabled={disabled}
				onKeyDown={onKeyDown}
				onChange={(e) => {
					setDraft(e.target.value);
					const n = Number(e.target.value);
					if (e.target.value !== "" && Number.isInteger(n) && n >= 0 && n <= 100) onChange(n);
				}}
				onBlur={() => setDraft(String(value))}
				className="min-h-11 w-full rounded-md border border-border bg-bg px-2 text-right font-mono text-base tabular-nums"
			/>
		</div>
	);
}
