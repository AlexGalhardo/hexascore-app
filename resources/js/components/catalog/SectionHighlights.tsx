import type { Cover } from "../CoverThumb";

// Shapes of the home highlights props (app/catalog SectionHighlights).
export interface Highlight {
	id: string;
	type: string;
	slug: string;
	title: string;
	year: number | null;
	score: number | null;
	ratingCount: number;
	cover: Cover | null;
}

export interface Section {
	type: "game" | "movie" | "series" | "book";
	count: number;
	topRated: Highlight[];
	newest: Highlight[];
	mostRated: Highlight[];
}

export const SECTION_PATHS: Record<Section["type"], string> = {
	game: "games",
	movie: "movies",
	series: "series",
	book: "books",
};
