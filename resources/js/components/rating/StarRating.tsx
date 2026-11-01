import { usePage } from "@inertiajs/react";
import { type KeyboardEvent, useId, useState } from "react";

import { useT } from "../../lib/useT";
import type { SharedProps } from "../../types";
import { formatScore } from "../ui/ScoreBadge";

const MAX = 10;
const STEP = 0.5;

const clamp = (v: number): number => Math.min(MAX, Math.max(STEP, v));

const STAR_PATH = "M12 2.5l2.9 6.3 6.9.8-5.1 4.7 1.4 6.8L12 17.6 5.9 21.1l1.4-6.8L2.2 9.6l6.9-.8z";

type StarFill = "full" | "half" | "empty";

// One shared symbol keeps the 120 stars of a rating panel small in the SSR HTML; render it once
// per page (RatingPanel does). The filled layer is clipped to the left half for a half star.
export function StarSymbol() {
	return (
		<svg aria-hidden="true" className="absolute size-0">
			<symbol id="hs-star" viewBox="0 0 24 24">
				<path d={STAR_PATH} strokeWidth="1.5" strokeLinejoin="round" />
			</symbol>
		</svg>
	);
}

function Star({ fill }: { fill: StarFill }) {
	return (
		<svg viewBox="0 0 24 24" aria-hidden="true" data-fill={fill} className="size-6 shrink-0">
			<use href="#hs-star" className="fill-none stroke-field" />
			{fill !== "empty" && (
				<use
					href="#hs-star"
					className={`fill-star stroke-star-line${fill === "half" ? " [clip-path:inset(0_50%_0_0)]" : ""}`}
				/>
			)}
		</svg>
	);
}

export interface StarRatingProps {
	label: string;
	value: number | null;
	onChange: (value: number | null) => void;
	disabled?: boolean;
	/** Id of the (visually hidden) help text that explains the keys. */
	describedBy?: string;
}

const KEY_MOVES: Record<string, (current: number) => number | null> = {
	ArrowRight: (c) => clamp(c + STEP),
	ArrowUp: (c) => clamp(c + STEP),
	ArrowLeft: (c) => clamp(c - STEP),
	ArrowDown: (c) => clamp(c - STEP),
	Home: () => STEP,
	End: () => MAX,
	Delete: () => null,
	Backspace: () => null,
};

// Ten stars, each two native radios (left half n - 0.5, right half n): one radio group of
// 20 options per category (design v1 rating-widget/states). Arrow keys clamp instead of
// wrapping, Home/End jump to the ends, Delete clears; the value is printed beside the label.
export default function StarRating({
	label,
	value,
	onChange,
	disabled = false,
	describedBy,
}: StarRatingProps) {
	const t = useT();
	const { locale } = usePage<SharedProps>().props;
	const name = useId();
	const [hover, setHover] = useState<number | null>(null);
	const shown = hover ?? value ?? 0;

	const onKeyDown = (e: KeyboardEvent<HTMLDivElement>) => {
		const move = KEY_MOVES[e.key];
		if (!move || disabled) return;
		e.preventDefault();
		const next = move(value ?? 0);
		onChange(next);
		if (next !== null) {
			e.currentTarget.querySelector<HTMLInputElement>(`input[value="${next}"]`)?.focus();
		}
	};

	const status =
		hover !== null && hover !== value
			? t("rating.preview", { value: formatScore(hover, locale) })
			: value === null
				? t("rating.unset")
				: t("rating.outOf", { value: formatScore(value, locale) });

	return (
		<div className="grid gap-1 border-b border-border pb-2">
			<div className="flex items-baseline justify-between gap-2">
				<span id={`${name}-label`} className="text-sm font-semibold">
					{label}
				</span>
				<span
					data-testid="rating-value"
					className={`text-sm tabular-nums ${value === null && hover === null ? "text-muted" : "font-semibold"}`}
				>
					{status}
				</span>
			</div>
			<div
				role="radiogroup"
				aria-labelledby={`${name}-label`}
				aria-describedby={describedBy}
				aria-disabled={disabled || undefined}
				className="flex w-fit cursor-pointer rounded-xs outline-offset-2 has-[:focus-visible]:outline-2 has-[:focus-visible]:outline-primary aria-disabled:cursor-not-allowed"
				onKeyDown={onKeyDown}
				onPointerLeave={() => setHover(null)}
			>
				{Array.from({ length: MAX }, (_, i) => (
					<span
						// biome-ignore lint/suspicious/noArrayIndexKey: fixed list of 10 positions
						key={i}
						className="relative flex h-11 w-7 items-center justify-center"
					>
						<Star
							fill={shown >= i + 1 ? "full" : shown >= i + 0.5 ? "half" : "empty"}
						/>
						{[i + 0.5, i + 1].map((v) => (
							<label
								key={v}
								className={`absolute inset-y-0 w-1/2 ${v % 1 ? "left-0" : "right-0"}`}
								onPointerEnter={() => !disabled && setHover(v)}
							>
								<input
									type="radio"
									name={name}
									value={v}
									checked={value === v}
									disabled={disabled}
									onChange={() => onChange(v)}
									aria-label={t("rating.valuetext", { value: v })}
									className="sr-only"
								/>
							</label>
						))}
					</span>
				))}
			</div>
		</div>
	);
}
