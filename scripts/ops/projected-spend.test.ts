import { expect, test } from "bun:test";

import { projectedUsd, shouldAlert } from "./projected-spend.ts";

test("day 20 with US$19 in a 30-day cycle projects US$28.5 and alerts", () => {
	expect(projectedUsd(19, 20, 30)).toBe(28.5);
	expect(shouldAlert(19, 20, 30)).toBe(true);
});

test("day 20 with US$17 in a 30-day cycle does not alert", () => {
	expect(projectedUsd(17, 20, 30)).toBe(25.5);
	expect(shouldAlert(17, 20, 30)).toBe(false);
});

test("a projection above US$27 before day 20 does not alert", () => {
	expect(shouldAlert(19, 19, 30)).toBe(false);
});
