import { expect, test } from "vitest";

import { abbreviateLabel, pageWindow } from "./pageWindow";

test("pageWindow keeps the edges, the neighbours, and a single hidden page", () => {
	expect(pageWindow(1, 1)).toEqual([1]);
	expect(pageWindow(3, 5)).toEqual([1, 2, 3, 4, 5]);
	expect(pageWindow(1, 5)).toEqual([1, 2, "gap", 5]);
	expect(pageWindow(5, 10)).toEqual([1, "gap", 4, 5, 6, "gap", 10]);
});

test("abbreviateLabel uses initials for several words and four characters otherwise", () => {
	expect(abbreviateLabel("Qualidade sonora")).toBe("QS");
	expect(abbreviateLabel("Trilha/Som")).toBe("TS");
	expect(abbreviateLabel("Gameplay")).toBe("Game");
});
