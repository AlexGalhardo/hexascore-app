import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen, waitFor } from "@testing-library/react-native";
import type { TrendingEntryJSON, TrendingJSON } from "../../../../packages/app-client/src/api";
import { HomeBody, type TrendingReader } from "./Home";

function entry(type: string, rank: number, basis: "trend" | "prior" = "trend"): TrendingEntryJSON {
	return {
		rank,
		basis,
		item: {
			id: `${type}-${rank}`,
			type,
			slug: `${type}-${rank}`,
			locale: "en",
			title: `${type} ${rank}`,
			score: 8,
		},
	};
}

function reader(
	byType: Partial<Record<"games" | "movies" | "series" | "books", TrendingEntryJSON[]>>,
): TrendingReader {
	const page = (data: TrendingEntryJSON[] | undefined): Promise<TrendingJSON> =>
		Promise.resolve({ computedAt: null, data: data ?? [] });
	return {
		listGameTrending: jest.fn(() => page(byType.games)),
		listMovieTrending: jest.fn(() => page(byType.movies)),
		listSeriesTrending: jest.fn(() => page(byType.series)),
		listBookTrending: jest.fn(() => page(byType.books)),
	};
}

async function renderHome(read: TrendingReader, locale: "en" | "pt-BR" = "en"): Promise<void> {
	const client = new QueryClient({
		defaultOptions: { queries: { retry: false, gcTime: Number.POSITIVE_INFINITY } },
	});
	await render(
		<QueryClientProvider client={client}>
			<HomeBody read={read} locale={locale} />
		</QueryClientProvider>,
	);
}

test("sections follow ITEM_TYPES order and an empty type is hidden", async () => {
	await renderHome(
		reader({
			books: [entry("books", 1)],
			games: [entry("games", 1)],
			series: [entry("series", 1)],
		}),
	);
	const titles = (): unknown[] =>
		screen.queryAllByTestId("section-title").map((node) => node.props.children);
	await waitFor(() => expect(titles()).toHaveLength(3));
	expect(titles()).toEqual([
		"Games: top 10 trending (30 days)",
		"TV series: top 10 trending (30 days)",
		"Books: top 10 trending (30 days)",
	]);
	expect(screen.queryByText(/Movies/)).toBeNull();
});

test("a prior entry shows the Popular chip and keeps the trending title", async () => {
	await renderHome(reader({ games: [entry("games", 1), entry("games", 2, "prior")] }));
	expect(await screen.findByText("Games: top 10 trending (30 days)")).toBeTruthy();
	expect(screen.getAllByText("Popular")).toHaveLength(1);
});

test("when every entry is prior the title becomes Popular {type}", async () => {
	await renderHome(
		reader({ movies: [entry("movies", 1, "prior"), entry("movies", 2, "prior")] }),
		"pt-BR",
	);
	expect(await screen.findByText("Filmes populares")).toBeTruthy();
});

test("every row prints its score next to the band colour", async () => {
	await renderHome(reader({ games: [entry("games", 1), entry("games", 2)] }));
	await screen.findByText("games 1");
	expect(screen.getAllByTestId("score-band-score-high")).toHaveLength(2);
	expect(screen.getAllByText("8.0")).toHaveLength(2);
});

test("no trending data anywhere shows the empty state, never made-up rows", async () => {
	await renderHome(reader({}));
	expect(await screen.findByText("Nothing is trending yet.")).toBeTruthy();
	expect(screen.queryAllByTestId("section-title")).toHaveLength(0);
});

test("a failed type is hidden like an empty one", async () => {
	const read = reader({ games: [entry("games", 1)] });
	read.listMovieTrending = jest.fn(() => Promise.reject(new Error("offline")));
	await renderHome(read);
	expect(await screen.findByText("games 1")).toBeTruthy();
	expect(screen.queryByText(/Movies/)).toBeNull();
});
