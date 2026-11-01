import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { fireEvent, render, screen, waitFor } from "@testing-library/react-native";
import type {
	AppSearchJSON,
	AppTrendingJSON,
	SearchHit,
	SearchIndexQuery,
	SearchSuggestQuery,
	SuggestJSON,
} from "../../../../packages/app-client/src/api";
import { ApiError } from "../../../../packages/app-client/src/api";
import { SearchBody, type SearchReader } from "../screens/Search";
import { type SearchClick, searchClickSender } from "./events";
import { cutQuery, MAX_QUERY } from "./useSuggest";

function hits(prefix: string, count: number, type = "game"): SearchHit[] {
	return Array.from({ length: count }, (_, i) => ({
		id: `${prefix}${i}`,
		type,
		slug: `${prefix}-slug-${i}`,
		title: `${prefix} ${i}`,
		path: `/games/${prefix}-slug-${i}`,
		cover: { thumb: `https://cdn.test/${prefix}${i}.webp`, full: "" },
		year: i === 0 ? null : 2015,
		score: i === 0 ? null : 9.1,
	}));
}

function suggestions(prefix: string, count: number, type = "game"): SuggestJSON {
	const section = { game: "games", movie: "movies", series: "series", book: "books" }[type];
	return {
		data: Array.from({ length: count }, (_, i) => ({
			id: `${prefix}${i}`,
			type,
			kind: type,
			title: `${prefix} ${i}`,
			year: 2015,
			coverThumbUrl: `https://cdn.test/${prefix}${i}.webp`,
			path: `/${section}/${prefix}-slug-${i}`,
		})),
	};
}

type Trending = Partial<Record<"game" | "movie" | "series" | "book", AppTrendingJSON | Error>>;

function trendingFor(type: keyof Trending, trending: Trending): jest.Mock {
	return jest.fn(() => {
		const out = trending[type] ?? { data: [] };
		return out instanceof Error ? Promise.reject(out) : Promise.resolve(out);
	});
}

function entry(type: string, rank: number, title: string) {
	return {
		basis: "trend" as const,
		rank,
		item: {
			id: `${type}-${rank}`,
			type,
			locale: "en",
			slug: `${type}-slug-${rank}`,
			title,
			releaseDate: "2020-05-01",
			coverThumbUrl: `https://cdn.test/${type}${rank}.webp`,
			score: 8.4,
		},
	};
}

function settle<T>(out: T | Error): Promise<T> {
	return out instanceof Error ? Promise.reject(out) : Promise.resolve(out);
}

function reader(
	answer: (query: SearchIndexQuery) => AppSearchJSON | Error,
	suggest: (query: SearchSuggestQuery) => SuggestJSON | Error = (q) =>
		suggestions(q.q?.slice(0, 1) ?? "s", 3),
	trending: Trending = {},
): SearchReader & { calls: SearchIndexQuery[]; suggestCalls: SearchSuggestQuery[] } {
	const calls: SearchIndexQuery[] = [];
	const suggestCalls: SearchSuggestQuery[] = [];
	return {
		calls,
		suggestCalls,
		listGameTrending: trendingFor("game", trending),
		listMovieTrending: trendingFor("movie", trending),
		listSeriesTrending: trendingFor("series", trending),
		listBookTrending: trendingFor("book", trending),
		searchIndex: jest.fn(({ query } = {}) => {
			calls.push(query ?? {});
			return settle(answer(query ?? {}));
		}),
		searchSuggest: jest.fn(({ query } = {}) => {
			suggestCalls.push(query ?? {});
			return settle(suggest(query ?? {}));
		}),
	};
}

async function renderSearch(
	read: SearchReader,
	delay = 0,
): Promise<{ onOpen: jest.Mock; onClick: jest.Mock<void, [SearchClick]> }> {
	const client = new QueryClient({
		defaultOptions: { queries: { retry: false, gcTime: Number.POSITIVE_INFINITY } },
	});
	const onOpen = jest.fn();
	const onClick = jest.fn<void, [SearchClick]>();
	await render(
		<QueryClientProvider client={client}>
			<SearchBody
				read={read}
				locale="en"
				delay={delay}
				onOpen={onOpen}
				onCancel={jest.fn()}
				onClick={onClick}
			/>
		</QueryClientProvider>,
	);
	return { onOpen, onClick };
}

const input = () => screen.getByLabelText("Search games, movies, series, books…");
const submit = async (): Promise<void> => {
	await fireEvent(input(), "submitEditing");
};

test("typing suggests with q and locale, and a suggestion opens the item route", async () => {
	const read = reader(() => ({ searchId: "s", data: [], total: 0 }));
	const { onOpen, onClick } = await renderSearch(read);
	await fireEvent.changeText(input(), "witcher");
	expect(await screen.findByText("w 1")).toBeTruthy();
	expect(read.suggestCalls.at(-1)).toEqual({ q: "witcher", locale: "en" });
	expect(read.calls).toHaveLength(0);
	expect(screen.getByText("Suggestions")).toBeTruthy();
	// Suggestions carry no score, so the row does not claim "not enough ratings".
	await fireEvent.press(screen.getByRole("button", { name: "w 1, 2015, Game" }));
	expect(onOpen).toHaveBeenCalledWith("/games/w-slug-1?id=w1");
	expect(onClick).not.toHaveBeenCalled();
});

test("one suggest call per debounce window", async () => {
	const read = reader(() => ({ searchId: "s", data: [], total: 0 }));
	await renderSearch(read, 50);
	for (const text of ["wi", "wit", "witc", "witch"]) await fireEvent.changeText(input(), text);
	expect(await screen.findByText("w 0")).toBeTruthy();
	expect(read.suggestCalls).toEqual([{ q: "witch", locale: "en" }]);

	await fireEvent.changeText(input(), "witche");
	await fireEvent.changeText(input(), "witcher");
	await waitFor(() => expect(read.suggestCalls).toHaveLength(2));
	expect(read.suggestCalls[1]).toEqual({ q: "witcher", locale: "en" });
});

test("fewer than 2 characters calls nothing", async () => {
	const read = reader(() => ({ searchId: "s", data: [], total: 0 }));
	await renderSearch(read);
	await fireEvent.changeText(input(), "w");
	await submit();
	await waitFor(() => expect(screen.queryByText("Searching…")).toBeNull());
	expect(read.suggestCalls).toHaveLength(0);
	expect(read.calls).toHaveLength(0);
});

test(`q over ${MAX_QUERY} characters is cut before it is sent`, async () => {
	const long = `  ${"é".repeat(60)}${"x".repeat(60)}  `;
	const cut = `${"é".repeat(60)}${"x".repeat(40)}`;
	expect(cutQuery(long)).toBe(cut);
	expect([...cutQuery("😀".repeat(150))]).toHaveLength(MAX_QUERY);

	const read = reader(() => ({ searchId: "s", data: hits("w", 1), total: 1 }));
	await renderSearch(read);
	await fireEvent.changeText(input(), long);
	await waitFor(() => expect(read.suggestCalls).toHaveLength(1));
	expect(read.suggestCalls[0]?.q).toBe(cut);
	await submit();
	await waitFor(() => expect(read.calls).toHaveLength(1));
	expect(read.calls[0]?.q).toBe(cut);
});

test("a type chip narrows the suggestions and the search", async () => {
	const read = reader(
		() => ({ searchId: "s", data: hits("m", 1, "movie"), total: 1 }),
		(q) => (q.type === "movie" ? suggestions("m", 1, "movie") : suggestions("w", 2)),
	);
	await renderSearch(read);
	await fireEvent.changeText(input(), "witcher");
	await screen.findByText("w 0");
	await fireEvent.press(screen.getByRole("button", { name: "Movies" }));
	expect(await screen.findByText("m 0")).toBeTruthy();
	expect(read.suggestCalls.at(-1)).toEqual({ q: "witcher", type: "movie", locale: "en" });
	await submit();
	await waitFor(() => expect(read.calls).toHaveLength(1));
	expect(read.calls[0]).toEqual({ q: "witcher", type: "movie", locale: "en", page: "1" });
	expect(screen.getByRole("button", { name: "Movies" }).props.accessibilityState).toEqual(
		expect.objectContaining({ selected: true }),
	);
});

test("submit runs the full search, grouped by type, and Next loads page 2", async () => {
	const read = reader((q) =>
		q.page === "2"
			? { searchId: "s2", data: hits("b", 2, "book"), total: 7 }
			: { searchId: "s1", data: [...hits("g", 3), ...hits("m", 2, "movie")], total: 7 },
	);
	await renderSearch(read);
	await fireEvent.changeText(input(), "an");
	await screen.findByText("Suggestions");
	await fireEvent.press(screen.getByRole("button", { name: "See all results for “an”" }));
	expect(await screen.findByText("7 results")).toBeTruthy();
	expect(read.calls.at(-1)).toEqual({ q: "an", locale: "en", page: "1" });
	expect(screen.getByRole("header", { name: "Games" })).toBeTruthy();
	expect(screen.getByRole("header", { name: "Movies" })).toBeTruthy();
	expect(screen.getByRole("button", { name: "g 1, 2015, Game, Score 9.1 out of 10" })).toBeTruthy();
	expect(screen.getByRole("button", { name: "g 0, Game, Not enough ratings yet" })).toBeTruthy();
	await fireEvent.press(screen.getByRole("button", { name: "Next" }));
	expect(await screen.findByRole("header", { name: "Books" })).toBeTruthy();
	expect(read.calls.at(-1)).toEqual({ q: "an", locale: "en", page: "2" });
	expect(screen.queryByRole("button", { name: "Next" })).toBeNull();
});

test("a result click sends the search id, item id, position in its group and group, never the query", async () => {
	const read = reader((q) =>
		q.page === "2"
			? { searchId: "s2", data: hits("n", 1, "movie"), total: 6 }
			: { searchId: "s1", data: [...hits("g", 3), ...hits("m", 2, "movie")], total: 6 },
	);
	const { onOpen, onClick } = await renderSearch(read);
	await fireEvent.changeText(input(), "secret query");
	await screen.findByText("Suggestions");
	await submit();
	await fireEvent.press(await screen.findByRole("button", { name: /^m 1,/ }));
	expect(onClick).toHaveBeenLastCalledWith({
		searchId: "s1",
		refId: "m1",
		position: 2,
		group: "movie",
	});
	expect(onOpen).toHaveBeenLastCalledWith("/movies/m-slug-1?id=m1");

	await fireEvent.press(screen.getByRole("button", { name: "Next" }));
	await fireEvent.press(await screen.findByRole("button", { name: /^n 0,/ }));
	expect(onClick).toHaveBeenLastCalledWith({
		searchId: "s2",
		refId: "n0",
		position: 3,
		group: "movie",
	});
	expect(JSON.stringify(onClick.mock.calls)).not.toContain("secret");
});

test("a failed suggest or search says so instead of no results", async () => {
	const busy = new ApiError(503, { status: 503, title: "busy" });
	const read = reader(
		() => busy,
		() => busy,
	);
	await renderSearch(read);
	await fireEvent.changeText(input(), "witcher");
	const message = "Search is busy right now. Please try again in a few seconds.";
	expect(await screen.findByText(message)).toBeTruthy();
	await submit();
	await waitFor(() => expect(read.calls).toHaveLength(1));
	expect(await screen.findByText(message)).toBeTruthy();
});

test("an opened suggestion becomes a recent search", async () => {
	const read = reader(() => ({ searchId: "s", data: [], total: 0 }));
	await renderSearch(read);
	await fireEvent.changeText(input(), "witcher");
	await fireEvent.press(await screen.findByRole("button", { name: /^w 0,/ }));
	await fireEvent.changeText(input(), "");
	expect(await screen.findByText("Recent searches")).toBeTruthy();
	expect(
		screen.getByRole("button", { name: "Remove “witcher” from recent searches" }),
	).toBeTruthy();
	await fireEvent.press(screen.getByRole("button", { name: "witcher" }));
	expect(input().props.value).toBe("witcher");
});

test("the empty sheet shows the trending empty state", async () => {
	await renderSearch(reader(() => ({ searchId: "s", data: [], total: 0 })));
	expect(await screen.findByText("Nothing is trending this week yet.")).toBeTruthy();
});

test("while trending loads the sheet shows a loading state, not the empty message", async () => {
	const read = {
		...reader(() => ({ searchId: "s", data: [], total: 0 })),
		listGameTrending: jest.fn(() => new Promise<never>(() => {})),
	};
	await renderSearch(read);
	expect(await screen.findByText("Loading…")).toBeTruthy();
	expect(screen.queryByText("Nothing is trending this week yet.")).toBeNull();
});

test("trending lists the top items across types by rank and a row opens the item", async () => {
	const read = reader(() => ({ searchId: "s", data: [], total: 0 }), undefined, {
		game: { data: [entry("game", 1, "Hades"), entry("game", 2, "Celeste")] },
		movie: { data: [entry("movie", 1, "Dune")] },
		series: new Error("down"),
		book: { data: [entry("book", 3, "Emma")] },
	});
	const { onOpen, onClick } = await renderSearch(read);
	expect(await screen.findByText("Hades")).toBeTruthy();
	const titles = screen
		.getAllByRole("button", { name: /, 2020, / })
		.map((row) => row.props.accessibilityLabel ?? "");
	expect(titles.map((name: string) => name.split(",")[0])).toEqual([
		"Hades",
		"Dune",
		"Celeste",
		"Emma",
	]);
	expect(screen.queryByText("Nothing is trending this week yet.")).toBeNull();
	expect(read.listMovieTrending).toHaveBeenCalledWith({ query: { locale: "en" } });
	await fireEvent.press(screen.getByRole("button", { name: /^Dune, 2020, Movie,/ }));
	expect(onOpen).toHaveBeenCalledWith("/movies/movie-slug-1?id=movie-1");
	expect(onClick).not.toHaveBeenCalled();
});

test("the click sender posts the body and swallows a failed post", async () => {
	const post = jest.fn(() => Promise.reject(new Error("offline")));
	const click: SearchClick = { searchId: "s", refId: "i", position: 1, group: "game" };
	searchClickSender(post)(click);
	await Promise.resolve();
	expect(post).toHaveBeenCalledWith(click);
});
