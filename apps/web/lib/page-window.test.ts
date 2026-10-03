import { describe, expect, test } from "bun:test";
import { pageWindow } from "./page-window";

describe("pageWindow", () => {
	test("compacts long ranges around the current page", () => {
		expect(pageWindow(5, 100)).toEqual([1, "gap", 4, 5, 6, "gap", 100]);
		expect(pageWindow(1, 100)).toEqual([1, 2, "gap", 100]);
		expect(pageWindow(100, 100)).toEqual([1, "gap", 99, 100]);
		expect(pageWindow(3, 100)).toEqual([1, 2, 3, 4, "gap", 100]);
	});

	test("short ranges and edge counts", () => {
		expect(pageWindow(1, 0)).toEqual([]);
		expect(pageWindow(1, 1)).toEqual([1]);
		expect(pageWindow(2, 5)).toEqual([1, 2, 3, 4, 5]);
	});
});
