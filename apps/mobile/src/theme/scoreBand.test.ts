import { scoreBand } from "./scoreBand";

describe("score bands", () => {
	test.each([
		[null, "score-none"],
		[0, "score-low"],
		[4.9, "score-low"],
		[5, "score-mid"],
		[7.4, "score-mid"],
		[7.5, "score-high"],
		[10, "score-high"],
	] as const)("%s → %s", (score, band) => {
		expect(scoreBand(score)).toBe(band);
	});
});
