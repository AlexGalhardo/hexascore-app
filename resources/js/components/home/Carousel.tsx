import { type ReactElement, type ReactNode, useRef, useState } from "react";

import { useT } from "../../lib/useT";
import { iconButton } from "../Nav";

const raised = `${iconButton} border border-border bg-surface`;

// Full class names so Tailwind keeps them; the landing carousel has at most 4 slides (12 items, 3 per slide).
const OFFSET = ["translate-x-0", "-translate-x-full", "-translate-x-[200%]", "-translate-x-[300%]"];

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
 * Manual carousel: previous/next buttons, labelled indicators, arrow keys and touch swipe.
 * Inactive slides are inert and hidden from assistive tech.
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

	const go = (i: number) => setIndex((i + count) % count);

	return (
		<section
			aria-roledescription="carousel"
			aria-label={label}
			className="grid gap-3"
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
			<div className="overflow-hidden">
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
				<div className="flex items-center justify-center gap-1">
					<button
						type="button"
						className={raised}
						aria-label={t("home.carousel.previous")}
						onClick={() => go(index - 1)}
					>
						<Glyph name="previous" />
					</button>
					{slides.slice(0, count).map((slide, i) => (
						<button
							key={slide.key}
							type="button"
							className={`${iconButton} group`}
							aria-label={t("home.carousel.goTo", { n: i + 1 })}
							aria-current={i === index ? "true" : undefined}
							onClick={() => go(i)}
						>
							<span className="size-2.5 rounded-full border border-fg group-aria-[current=true]:bg-fg" />
						</button>
					))}
					<button
						type="button"
						className={raised}
						aria-label={t("home.carousel.next")}
						onClick={() => go(index + 1)}
					>
						<Glyph name="next" />
					</button>
				</div>
			)}
		</section>
	);
}
