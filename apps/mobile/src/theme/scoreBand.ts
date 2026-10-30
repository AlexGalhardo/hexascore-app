export type ScoreBand = "score-low" | "score-mid" | "score-high" | "score-none";

// SDD 10 §7.2. Half-star scores land on one side of 7.5; 7.4 is still mid.
export function scoreBand(score: number | null): ScoreBand {
	if (score === null) return "score-none";
	if (score < 5) return "score-low";
	if (score < 7.5) return "score-mid";
	return "score-high";
}
