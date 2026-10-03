import { describe, expect, test } from "bun:test";
import { abbreviateLabel, CATEGORY_KEYS, hasEnoughRatings, overall, scoreBand } from "./categories";

describe("scoreBand", () => {
	test.each([
		[null, "none"],
		[undefined, "none"],
		[0, "low"],
		[49, "low"],
		[50, "mid"],
		[74, "mid"],
		[75, "high"],
		[100, "high"],
	] as const)("%p -> %s", (score, band) => {
		expect(scoreBand(score)).toBe(band);
	});
});

describe("overall", () => {
	test("rounds the mean half up like Postgres round()", () => {
		expect(overall(80, 80, 80, 80, 80, 80)).toBe(80);
		expect(overall(0, 0, 0, 0, 0, 3)).toBe(1);
		expect(overall(100, 100, 100, 100, 100, 99)).toBe(100);
		expect(overall(50, 50, 50, 50, 50, 53)).toBe(51);
	});

	test("requires exactly 6 values", () => {
		expect(() => overall(1, 2, 3)).toThrow();
	});
});

test("abbreviateLabel", () => {
	expect(abbreviateLabel("Gameplay")).toBe("Game");
	expect(abbreviateLabel("Qualidade sonora")).toBe("QS");
	expect(abbreviateLabel("Trilha/Som")).toBe("TS");
});

test("metascore threshold is 5 ratings", () => {
	expect(hasEnoughRatings(4)).toBe(false);
	expect(hasEnoughRatings(5)).toBe(true);
});

test("every type has 6 unique category keys", () => {
	for (const keys of Object.values(CATEGORY_KEYS)) {
		expect(new Set(keys).size).toBe(6);
	}
});
