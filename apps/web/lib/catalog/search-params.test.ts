import { expect, test } from "bun:test";
import { parseCatalogSearchParams, sortMessageKey } from "./search-params";

test("parses filters and maps category names to columns", () => {
	const { params, query } = parseCatalogSearchParams("game", {
		genre: "RPG",
		year: "2020",
		min_score: "70",
		min_gameplay: "80",
		min_storytelling: "60",
		sort: "-score",
		cursor: "abc_-",
	});
	expect(params).toMatchObject({
		genre: "RPG",
		year: 2020,
		minScore: 70,
		minCategory: { c1: 80, c6: 60 },
		sort: "-score",
		cursor: "abc_-",
	});
	expect(query).toEqual({
		genre: "RPG",
		year: "2020",
		min_score: "70",
		min_gameplay: "80",
		min_storytelling: "60",
		sort: "-score",
	});
});

test("ignores invalid values instead of failing", () => {
	const { params, query } = parseCatalogSearchParams("movie", {
		year: "soon",
		min_score: "101",
		sort: "chaos",
		platform: "PC",
		min_gameplay: "50",
		genre: ["Drama", "Crime"],
		cursor: "<script>",
	});
	expect(params).toMatchObject({ genre: "Drama", minCategory: {} });
	expect(params.year).toBeUndefined();
	expect(params.minScore).toBeUndefined();
	expect(params.sort).toBeUndefined();
	expect(params.platform).toBeUndefined();
	expect(params.cursor).toBeUndefined();
	expect(query).toEqual({ genre: "Drama" });
});

test("sort message keys", () => {
	expect(sortMessageKey("-release_date")).toBe("release_date_desc");
	expect(sortMessageKey("score")).toBe("score_asc");
});
