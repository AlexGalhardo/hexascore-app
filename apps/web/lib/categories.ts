import type { ItemType } from "./db/schema/catalog";

// Column order c1..c6 in `ratings`/`items.score_c*`. Keys double as public API filter names (`min_gameplay`)
// and as message keys in messages/<locale>/categories.json — renaming one is a breaking change.
export const CATEGORY_KEYS = {
	game: ["gameplay", "fun", "sound", "technical", "visual", "storytelling"],
	movie: ["screenplay", "acting", "direction", "cinematography", "soundtrack", "entertainment"],
	series: ["screenplay", "acting", "characters", "pacing", "production", "entertainment"],
	book: ["plot", "characters", "writing", "pacing", "originality", "impact"],
} as const satisfies Record<ItemType, readonly string[]>;

export type CategoryKey = (typeof CATEGORY_KEYS)[ItemType][number];

export const MIN_RATINGS_FOR_SCORE = 5;

export type ScoreBand = "low" | "mid" | "high" | "none";

export function scoreBand(score: number | null | undefined): ScoreBand {
	if (score === null || score === undefined || Number.isNaN(score)) return "none";
	if (score < 50) return "low";
	if (score < 75) return "mid";
	return "high";
}

export function overall(...categories: number[]): number {
	if (categories.length !== 6) throw new Error(`overall() expects 6 category scores, got ${categories.length}`);
	return Math.round(categories.reduce((sum, c) => sum + c, 0) / 6);
}

/** Short column header: "Gameplay" → "Game", "Qualidade sonora" → "QS", "Trilha/Som" → "TS". */
export function abbreviateLabel(label: string): string {
	const words = label.split(/[\s/]+/).filter(Boolean);
	return words.length > 1 ? words.map((w) => w.charAt(0).toUpperCase()).join("") : label.slice(0, 4);
}

export function hasEnoughRatings(ratingCount: number): boolean {
	return ratingCount >= MIN_RATINGS_FOR_SCORE;
}
