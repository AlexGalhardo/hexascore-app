import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { fireEvent, render, screen, waitFor } from "@testing-library/react-native";
import type {
	AppSearchJSON,
	AppTrendingJSON,
	SearchHit,
	SearchIndexQuery,
} from "../../../../packages/app-client/src/api";
import { ApiError } from "../../../../packages/app-client/src/api";
import { SearchBody, type SearchReader } from "./Search";

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

function reader(
	answer: (query: SearchIndexQuery) => AppSearchJSON | Error,
	trending: Trending = {},
): SearchReader & {
	calls: SearchIndexQuery[];
} {
	const calls: SearchIndexQuery[] = [];
	return {
		calls,
		listGameTrending: trendingFor("game", trending),
		listMovieTrending: trendingFor("movie", trending),
		listSeriesTrending: trendingFor("series", trending),
		listBookTrending: trendingFor("book", trending),
		searchIndex: jest.fn(({ query } = {}) => {
			calls.push(query ?? {});
			const out = answer(query ?? {});
			return out instanceof Error ? Promise.reject(out) : Promise.resolve(out);
		}),
	};
}

async function renderSearch(read: SearchReader): Promise<{ onOpen: jest.Mock }> {
	const client = new QueryClient({
		defaultOptions: { queries: { retry: false, gcTime: Number.POSITIVE_INFINITY } },
	});
	const onOpen = jest.fn();
	await render(
		<QueryClientProvider client={client}>
			<SearchBody read={read} locale="en" delay={0} onOpen={onOpen} onCancel={jest.fn()} />
		</QueryClientProvider>,
	);
	return { onOpen };
}

const input = () => screen.getByLabelText("Search games, movies, series, books…");

test("typing searches with q, locale and page 1, and a row opens the item route", async () => {
	const read = reader(() => ({ searchId: "s", data: hits("w", 3), total: 3 }));
	const { onOpen } = await renderSearch(read);
	await fireEvent.changeText(input(), "witcher");
	expect(await screen.findByText("w 0")).toBeTruthy();
	expect(read.calls.at(-1)).toEqual({ q: "witcher", locale: "en", page: "1" });
	expect(screen.getByText("3 results")).toBeTruthy();
	expect(screen.getByRole("button", { name: "w 1, 2015, Game, Score 9.1 out of 10" })).toBeTruthy();
	expect(screen.getByRole("button", { name: "w 0, Game, Not enough ratings yet" })).toBeTruthy();
	await fireEvent.press(screen.getByRole("button", { name: /^w 1, 2015, Game,/ }));
	expect(onOpen).toHaveBeenCalledWith("/games/w-slug-1?id=w1");
});

test("fewer than 2 characters does not call the API", async () => {
	const read = reader(() => ({ searchId: "s", data: [], total: 0 }));
	await renderSearch(read);
	await fireEvent.changeText(input(), "w");
	await waitFor(() => expect(screen.queryByText("Searching…")).toBeNull());
	expect(read.calls).toHaveLength(0);
});

test("a type chip narrows the search to that type", async () => {
	const read = reader((q) => ({
		searchId: "s",
		data: q.type === "movie" ? hits("m", 1, "movie") : hits("w", 2),
		total: q.type === "movie" ? 1 : 2,
	}));
	await renderSearch(read);
	await fireEvent.changeText(input(), "witcher");
	await screen.findByText("w 0");
	await fireEvent.press(screen.getByRole("button", { name: "Movies" }));
	expect(await screen.findByText("m 0")).toBeTruthy();
	expect(read.calls.at(-1)).toEqual({ q: "witcher", type: "movie", locale: "en", page: "1" });
	expect(screen.getByRole("button", { name: "Movies" }).props.accessibilityState).toEqual(
		expect.objectContaining({ selected: true }),
	);
});

test("see all lists the whole page and Next loads page 2", async () => {
	const read = reader((q) =>
		q.page === "2"
			? { searchId: "s", data: hits("b", 5), total: 25 }
			: { searchId: "s", data: hits("a", 20), total: 25 },
	);
	await renderSearch(read);
	await fireEvent.changeText(input(), "an");
	await screen.findByText("a 0");
	expect(screen.queryByText("a 8")).toBeNull();
	await fireEvent.press(screen.getByRole("button", { name: "See all results for “an”" }));
	expect(screen.getByText("a 19")).toBeTruthy();
	await fireEvent.press(screen.getByRole("button", { name: "Next" }));
	expect(await screen.findByText("b 4")).toBeTruthy();
	expect(read.calls.at(-1)).toEqual({ q: "an", locale: "en", page: "2" });
	// All 25 loaded: nothing left to fetch.
	expect(screen.queryByRole("button", { name: "Next" })).toBeNull();
});

test("a failed search says so instead of no results", async () => {
	const read = reader(() => new ApiError(503, { status: 503, title: "busy" }));
	await renderSearch(read);
	await fireEvent.changeText(input(), "witcher");
	expect(
		await screen.findByText("Search is busy right now. Please try again in a few seconds."),
	).toBeTruthy();
});

test("an opened search becomes a recent search", async () => {
	const read = reader(() => ({ searchId: "s", data: hits("w", 1), total: 1 }));
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
	const read = reader(() => ({ searchId: "s", data: [], total: 0 }), {
		game: { data: [entry("game", 1, "Hades"), entry("game", 2, "Celeste")] },
		movie: { data: [entry("movie", 1, "Dune")] },
		series: new Error("down"),
		book: { data: [entry("book", 3, "Emma")] },
	});
	const { onOpen } = await renderSearch(read);
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
});
