import { type ReactElement, type ReactNode, useRef, useState } from "react";

import { useT } from "../../lib/useT";

// Full class names so Tailwind keeps them; the landing carousel has at most 2 slides (12 items, 6 per slide).
const OFFSET = ["translate-x-0", "-translate-x-full", "-translate-x-[200%]", "-translate-x-[300%]"];

/**
 * Carousel without visible controls: navigation is by arrow keys and touch swipe.
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
		</section>
	);
}
