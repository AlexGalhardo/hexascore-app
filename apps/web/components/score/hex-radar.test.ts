import { expect, test } from "bun:test";
import { hexVertex, radarPoints } from "./hex-radar";

test("first axis points up and axes are 60° apart", () => {
	expect(hexVertex(0, 100)).toEqual([0, -100]);
	expect(hexVertex(3, 100)).toEqual([0, 100]);
	expect(hexVertex(1, 100)[0]).toBeCloseTo(86.6, 1);
});

test("radar points clamp to 0..100 and treat missing values as the centre", () => {
	expect(radarPoints([100, null, 150, -5, 50, 0]).split(" ")).toEqual([
		"0,-100",
		"0,0",
		"86.6,50",
		"0,0",
		"-43.3,25",
		"0,0",
	]);
});
