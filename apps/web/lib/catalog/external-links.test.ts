import { expect, test } from "bun:test";
import { buildExternalLinks } from "./external-links";

test("builds attributed score links and id-only links", () => {
	const links = buildExternalLinks("movie", { imdb: "tt0111161", rottentomatoes: "shawshank", hltb: "1" }, [
		{ source: "rottentomatoes", value: "91%", url: null },
		{ source: "imdb", value: "9.3", url: "javascript:alert(1)" },
	]);
	expect(links).toEqual([
		{
			provider: "rottentomatoes",
			source: "rottentomatoes",
			value: "91%",
			url: "https://www.rottentomatoes.com/m/shawshank",
		},
		{ provider: "imdb", source: "imdb", value: "9.3", url: "https://www.imdb.com/title/tt0111161/" },
		{ provider: "hltb", source: "hltb", value: null, url: "https://howlongtobeat.com/game/1" },
	]);
});

test("skips providers that do not apply to the type", () => {
	expect(buildExternalLinks("book", { metacritic: "x" }, [{ source: "metacritic", value: "80", url: null }])).toEqual(
		[],
	);
	expect(buildExternalLinks("series", { metacritic: "lost" }, [])[0]?.url).toBe(
		"https://www.metacritic.com/tv/lost/",
	);
});
