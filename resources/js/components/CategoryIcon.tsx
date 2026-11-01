import type { ReactElement } from "react";

export type CategoryType = "game" | "movie" | "series" | "book";

// Stroke icons on the 24 px grid of the navbar set: gamepad, clapperboard, television, open book.
const PATHS: Record<CategoryType, string> = {
	game: "M7 8 H17 a4 4 0 0 1 4 4 v1.5 a3 3 0 0 1-5.6 1.5 L14.5 13.5 H9.5 L8.6 15 A3 3 0 0 1 3 13.5 V12 a4 4 0 0 1 4-4 Z M8 10.5 V13 M6.75 11.75 H9.25 M15 11 h.01 M17 12.5 h.01",
	movie: "M4 9 H20 V19 H4 Z M4 9 L5 5 L19 5 L20 9 M9 5 L8 9 M14 5 L13 9",
	series: "M3 7 H21 V18 H3 Z M8 3 L12 7 L16 3 M8 21 H16",
	book: "M12 6.5 C10 5 7 4.5 3 5 V19 C7 18.5 10 19 12 20.5 C14 19 17 18.5 21 19 V5 C17 4.5 14 5 12 6.5 Z M12 6.5 V20.5",
};

/** Decorative icon placed before a category name; the name itself stays as text. */
export default function CategoryIcon({
	type,
	className = "size-5",
}: {
	type: CategoryType;
	className?: string;
}): ReactElement {
	return (
		<svg
			aria-hidden="true"
			data-category-icon={type}
			viewBox="0 0 24 24"
			className={`shrink-0 ${className}`}
			fill="none"
		>
			<path
				d={PATHS[type]}
				stroke="currentColor"
				strokeWidth="1.75"
				strokeLinecap="round"
				strokeLinejoin="round"
			/>
		</svg>
	);
}
