export type ItemType = "game" | "movie" | "series" | "book";

export interface Hours {
	mainHours: number | null;
	extrasHours: number | null;
	completionistHours: number | null;
}

export interface Me {
	rating: Record<string, number> | null;
	status: string | null;
	timeToBeat: Hours | null;
}

export interface RatingSummary {
	score: number | null;
	scoreByCategory: Record<string, number>;
	ratingCount: number;
	confidence: string | null;
	unusualActivity: boolean;
	ratingsLimitedUntil: string | null;
}

export interface EditionRow {
	id: string;
	title: string;
	subtitle: string | null;
	language: string;
	publisher: string | null;
	publicationYear: number | null;
	pageCount: number | null;
	format: string | null;
	isbn13: string | null;
}

export const CATEGORY_KEYS: Record<ItemType, string[]> = {
	game: ["gameplay", "fun", "sound", "technical", "visual", "storytelling"],
	movie: ["screenplay", "acting", "direction", "cinematography", "soundtrack", "entertainment"],
	series: ["screenplay", "acting", "characters", "pacing", "production", "entertainment"],
	book: ["plot", "characters", "writing", "pacing", "originality", "impact"],
};
