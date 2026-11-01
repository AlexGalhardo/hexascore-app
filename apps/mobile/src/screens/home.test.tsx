import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { fireEvent, render, screen, waitFor, within } from "@testing-library/react-native";
import type {
	AppTrendingEntryJSON,
	AppTrendingJSON,
} from "../../../../packages/app-client/src/api";
import { HomeBody, type TrendingReader } from "./Home";

function entry(
	type: string,
	rank: number,
	basis: "trend" | "prior" = "trend",
	score: number | null = 8,
): AppTrendingEntryJSON {
	return {
		rank,
		basis,
		item: {
			id: `${type}-${rank}`,
			type,
			slug: `${type}-slug-${rank}`,
			locale: "en",
			title: `${type} ${rank}`,
			score,
			ratingCount: 213,
			releaseDate: "2028-03-01",
		},
	};
}

function reader(
	byType: Partial<Record<"games" | "movies" | "series" | "books", AppTrendingEntryJSON[]>>,
): TrendingReader {
	const page = (data: AppTrendingEntryJSON[] | undefined): Promise<AppTrendingJSON> =>
		Promise.resolve({ computedAt: null, data: data ?? [] });
	return {
		listGameTrending: jest.fn(() => page(byType.games)),
		listMovieTrending: jest.fn(() => page(byType.movies)),
		listSeriesTrending: jest.fn(() => page(byType.series)),
		listBookTrending: jest.fn(() => page(byType.books)),
	};
}

async function renderHome(
	read: TrendingReader,
	locale: "en" | "pt-BR" = "en",
): Promise<{ onOpen: jest.Mock; onSeeAll: jest.Mock }> {
	const client = new QueryClient({
		defaultOptions: { queries: { retry: false, gcTime: Number.POSITIVE_INFINITY } },
	});
	const onOpen = jest.fn();
	const onSeeAll = jest.fn();
	await render(
		<QueryClientProvider client={client}>
			<HomeBody read={read} locale={locale} onOpen={onOpen} onSeeAll={onSeeAll} />
		</QueryClientProvider>,
	);
	return { onOpen, onSeeAll };
}

const sectionTitles = (): unknown[] =>
	screen
		.queryAllByRole("header")
		.map((node) => node.props.children)
		.filter((title) => title !== "HexaScore");

test("sections follow ITEM_TYPES order and an empty type is hidden", async () => {
	await renderHome(
		reader({
			books: [entry("books", 1)],
			games: [entry("games", 1)],
			series: [entry("series", 1)],
		}),
	);
	await waitFor(() => expect(sectionTitles()).toHaveLength(3));
	expect(sectionTitles()).toEqual([
		"Games: top 10 trending (30 days)",
		"TV series: top 10 trending (30 days)",
		"Books: top 10 trending (30 days)",
	]);
	expect(screen.queryByTestId("home-section-movie")).toBeNull();
});

test("a prior entry is labelled Popular and keeps the trending title", async () => {
	await renderHome(reader({ games: [entry("games", 1), entry("games", 2, "prior")] }));
	expect(await screen.findByText("Games: top 10 trending (30 days)")).toBeTruthy();
	expect(screen.getByText("#1 · Trending this week")).toBeTruthy();
	expect(screen.getByText("#2 · Popular")).toBeTruthy();
});

test("when every entry is prior the title becomes Popular {type}", async () => {
	await renderHome(
		reader({ movies: [entry("movies", 1, "prior"), entry("movies", 2, "prior")] }),
		"pt-BR",
	);
	expect(await screen.findByText("Filmes populares")).toBeTruthy();
});

test("three cards, then a grid of the next six; score, year and rating count shown", async () => {
	const games = Array.from({ length: 10 }, (_, index) => entry("games", index + 1));
	games[0] = entry("games", 1, "trend", null);
	await renderHome(reader({ games }));
	await screen.findByText("games 1");
	expect(screen.getByText("#3 · Trending this week")).toBeTruthy();
	expect(screen.queryByText("#4 · Trending this week")).toBeNull();
	expect(screen.getByText("games 9")).toBeTruthy();
	expect(screen.queryByText("games 10")).toBeNull();
	expect(screen.getAllByText("2028 · Games")).toHaveLength(3);
	expect(screen.getByText("Not enough ratings")).toBeTruthy();
	expect(screen.getAllByText("213 ratings")).toHaveLength(3);
	expect(screen.getAllByText("8.0")).toHaveLength(8);
	expect(screen.getByText("–")).toBeTruthy();
});

test("a card opens the API item and see-all opens the section", async () => {
	const { onOpen, onSeeAll } = await renderHome(reader({ books: [entry("books", 1)] }));
	await fireEvent.press(
		await screen.findByRole("button", { name: /^#1 · Trending this week, books 1,/ }),
	);
	expect(onOpen).toHaveBeenCalledWith(
		expect.objectContaining({ id: "books-1", slug: "books-slug-1" }),
	);
	await fireEvent.press(within(screen.getByTestId("home-section-book")).getByRole("link"));
	expect(onSeeAll).toHaveBeenCalledWith("books");
});

test("no trending data anywhere shows the empty state, never made-up rows", async () => {
	await renderHome(reader({}));
	expect(await screen.findByText("Nothing is trending yet.")).toBeTruthy();
	expect(sectionTitles()).toHaveLength(0);
});

test("a failed type is hidden like an empty one", async () => {
	const read = reader({ games: [entry("games", 1)] });
	read.listMovieTrending = jest.fn(() => Promise.reject(new Error("offline")));
	await renderHome(read);
	expect(await screen.findByText("games 1")).toBeTruthy();
	expect(screen.queryByText(/Movies/)).toBeNull();
});
