import { type ReactElement, type ReactNode, useEffect, useRef, useState } from "react";

import { useT } from "../../lib/useT";
import { iconButton } from "../Nav";

// Hidden until the carousel is hovered or focused, so the slide arrows do not clutter the grid.
const reveal =
	"border border-border bg-surface opacity-0 transition-opacity group-focus-within:opacity-100 group-hover:opacity-100 focus-visible:opacity-100 motion-reduce:transition-none";

// Full class names so Tailwind keeps them; the landing carousel has at most 2 slides (12 items, 6 per slide).
const OFFSET = ["translate-x-0", "-translate-x-full", "-translate-x-[200%]", "-translate-x-[300%]"];

const AUTO_MS = 6000;

const PATHS = {
	previous: "M15 6 L9 12 L15 18",
	next: "M9 6 L15 12 L9 18",
} as const;

function Glyph({ name }: { name: keyof typeof PATHS }): ReactElement {
	return (
		<svg aria-hidden="true" viewBox="0 0 24 24" className="size-5" fill="none">
			<path
				d={PATHS[name]}
				stroke="currentColor"
				strokeWidth="1.75"
				strokeLinecap="round"
				strokeLinejoin="round"
			/>
		</svg>
	);
}

/**
 * Carousel with arrow buttons that appear on hover or focus, plus arrow keys (the region is focusable)
 * and touch swipe. Inactive slides are inert and hidden from assistive tech.
 */
export default function Carousel({
	label,
	slides,
}: {
	label: string;
	slides: { key: string; node: ReactNode }[];
}): ReactElement {
	const t = useT();
	const count = Math.min(slides.length, OFFSET.length);
	const [index, setIndex] = useState(0);
	const touchX = useRef<number | null>(null);
	const [paused, setPaused] = useState(false);

	const go = (i: number) => setIndex((i + count) % count);

	// Pauses while hovered or focused so a visitor reading a card is not moved away; reduced motion disables it.
	useEffect(() => {
		if (count < 2 || paused) return;
		if (window.matchMedia?.("(prefers-reduced-motion: reduce)").matches) return;
		const id = window.setInterval(() => setIndex((i) => (i + 1) % count), AUTO_MS);
		return () => window.clearInterval(id);
	}, [count, paused]);

	return (
		<section
			aria-roledescription="carousel"
			aria-label={label}
			onMouseEnter={() => setPaused(true)}
			onMouseLeave={() => setPaused(false)}
			onFocus={() => setPaused(true)}
			onBlur={(e) => {
				if (!e.currentTarget.contains(e.relatedTarget as Node | null)) setPaused(false);
			}}
			tabIndex={0}
			className="group grid gap-3 outline-none focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fg"
			onKeyDown={(e) => {
				if (e.key === "ArrowRight") go(index + 1);
				else if (e.key === "ArrowLeft") go(index - 1);
			}}
			onTouchStart={(e) => {
				touchX.current = e.touches[0]?.clientX ?? null;
			}}
			onTouchEnd={(e) => {
				const start = touchX.current;
				const end = e.changedTouches[0]?.clientX;
				touchX.current = null;
				if (start === null || end === undefined || Math.abs(end - start) < 40) return;
				go(end < start ? index + 1 : index - 1);
			}}
		>
			<div className="relative overflow-hidden">
				<div
					aria-live="polite"
					className={`flex transition-transform duration-500 motion-reduce:transition-none ${OFFSET[index]}`}
				>
					{slides.slice(0, count).map((slide, i) => (
						// biome-ignore lint/a11y/useSemanticElements: APG slide pattern; fieldset is for form controls
						<div
							key={slide.key}
							role="group"
							aria-roledescription="slide"
							aria-label={t("home.carousel.slide", { n: i + 1, total: count })}
							aria-hidden={i !== index}
							inert={i !== index}
							className="w-full shrink-0"
						>
							{slide.node}
						</div>
					))}
				</div>
			</div>
			{count > 1 && (
				<>
					<button
						type="button"
						className={`${iconButton} ${reveal} absolute top-1/2 left-2 -translate-y-1/2`}
						aria-label={t("home.carousel.previous")}
						onClick={() => go(index - 1)}
					>
						<Glyph name="previous" />
					</button>
					<button
						type="button"
						className={`${iconButton} ${reveal} absolute top-1/2 right-2 -translate-y-1/2`}
						aria-label={t("home.carousel.next")}
						onClick={() => go(index + 1)}
					>
						<Glyph name="next" />
					</button>
				</>
			)}
		</section>
	);
}
