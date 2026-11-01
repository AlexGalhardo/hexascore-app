import { type ReactElement, type ReactNode, useEffect, useRef, useState } from "react";

import { useT } from "../../lib/useT";
import { iconButton } from "../Nav";

/** Time each slide stays before the next one while auto-advancing. */
export const SLIDE_MS = 4000;

// Full class names so Tailwind keeps them; the landing carousel has at most 4 slides (12 items, 3 per slide).
const OFFSET = ["translate-x-0", "-translate-x-full", "-translate-x-[200%]", "-translate-x-[300%]"];

const PATHS = {
	previous: "M15 6 L9 12 L15 18",
	next: "M9 6 L15 12 L9 18",
	pause: "M9 6 V18 M15 6 V18",
	play: "M8 5 L19 12 L8 19 Z",
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
 * Hand-written carousel (WAI-ARIA APG pattern, WCAG 2.2.2): auto-advances every SLIDE_MS unless the visitor paused
 * it, prefers reduced motion, hovers it or has focus inside; previous/next buttons, labelled indicators, arrow keys
 * and touch swipe; inactive slides are inert and hidden from assistive tech.
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
	// null = follow the motion preference; true/false = the visitor's explicit choice.
	const [choice, setChoice] = useState<boolean | null>(null);
	// Until the effect reads the media query (and on the server) nothing moves.
	const [reduced, setReduced] = useState(true);
	const [hovered, setHovered] = useState(false);
	const [focused, setFocused] = useState(false);
	const touchX = useRef<number | null>(null);

	useEffect(() => {
		const mq = window.matchMedia("(prefers-reduced-motion: reduce)");
		const sync = () => setReduced(mq.matches);
		sync();
		mq.addEventListener("change", sync);
		return () => mq.removeEventListener("change", sync);
	}, []);

	const playing = choice ?? !reduced;
	const auto = playing && !hovered && !focused && count > 1;

	useEffect(() => {
		if (!auto) return;
		const id = window.setInterval(() => setIndex((i) => (i + 1) % count), SLIDE_MS);
		return () => window.clearInterval(id);
	}, [auto, count]);

	const go = (i: number) => setIndex((i + count) % count);

	return (
		<section
			aria-roledescription="carousel"
			aria-label={label}
			className="grid gap-3"
			onMouseEnter={() => setHovered(true)}
			onMouseLeave={() => setHovered(false)}
			onFocus={() => setFocused(true)}
			onBlur={(e) => {
				if (!e.currentTarget.contains(e.relatedTarget)) setFocused(false);
			}}
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
					aria-live={auto ? "off" : "polite"}
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
						className={iconButton}
						aria-label={playing ? t("home.carousel.pause") : t("home.carousel.play")}
						onClick={() => setChoice(!playing)}
					>
						<Glyph name={playing ? "pause" : "play"} />
					</button>
					<button
						type="button"
						className={iconButton}
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
						className={iconButton}
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
