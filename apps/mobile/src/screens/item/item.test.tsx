import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen, waitFor } from "@testing-library/react-native";
import type { AppItemJSON, EditionPageJSON } from "../../../../../packages/app-client/src/api";
import { queryKeys } from "../../query/keys";
import { type ItemReader, ItemView } from "./Item";

const base: AppItemJSON = {
	id: "new",
	type: "books",
	slug: "dune",
	locale: "en",
	title: "Dune",
	score: 8.5,
	ratingCount: 12,
	confidence: "high",
	scoreByCategory: { plot: 9, characters: 8 },
	entities: [],
	provenance: { title: { source: "wikidata" } },
};

function reader(
	item: AppItemJSON,
	editions: EditionPageJSON = { data: [] },
	me: unknown = {},
): ItemReader {
	return {
		getItem: jest.fn(() => Promise.resolve(item)),
		listEditions: jest.fn(() => Promise.resolve(editions)),
		me: jest.fn(() => Promise.resolve(me)),
	};
}

async function renderItem(
	read: ItemReader,
	id = "new",
	onMerged?: (id: string) => void,
): Promise<QueryClient> {
	const client = new QueryClient({
		defaultOptions: { queries: { retry: false, gcTime: Number.POSITIVE_INFINITY } },
	});
	await render(
		<QueryClientProvider client={client}>
			<ItemView type="books" id={id} read={read} locale="en" onMerged={onMerged} />
		</QueryClientProvider>,
	);
	return client;
}

test("a merged id follows the 301 and rewrites the cache key", async () => {
	const read = reader(base);
	const onMerged = jest.fn();
	const client = await renderItem(read, "old", onMerged);
	expect(await screen.findByText("Dune")).toBeTruthy();
	await waitFor(() => expect(onMerged).toHaveBeenCalledWith("new"));
	expect(client.getQueryData(queryKeys.item("books", "new", "en"))).toMatchObject({ id: "new" });
	expect(client.getQueryData(queryKeys.item("books", "old", "en"))).toBeUndefined();
	await waitFor(() => expect(read.me).toHaveBeenCalledWith({ params: { id: "new" } }));
	expect(read.me).not.toHaveBeenCalledWith({ params: { id: "old" } });
});

test("empty edition and entity groups are hidden", async () => {
	await renderItem(reader(base));
	await screen.findByText("Dune");
	await waitFor(() => expect(screen.queryByText("Category averages")).toBeTruthy());
	expect(screen.queryByText("Editions")).toBeNull();
	expect(screen.queryByText("Author")).toBeNull();
});

test("entities are grouped by role and editions are listed when present", async () => {
	const item = {
		...base,
		entities: [
			{ id: "e1", role: "author", name: "Frank Herbert" },
			{ id: "e2", role: "publisher", name: "Chilton" },
			{ id: "e3", role: "author", name: "Second Author" },
		],
	};
	await renderItem(
		reader(item, { data: [{ id: "ed1", title: "Dune (1965)", format: "hardcover" }] }),
	);
	expect(await screen.findByText("Author")).toBeTruthy();
	expect(screen.getByText("Frank Herbert")).toBeTruthy();
	expect(screen.getByText("Second Author")).toBeTruthy();
	expect(screen.getAllByText("Author")).toHaveLength(1);
	expect(screen.getByText("Publisher")).toBeTruthy();
	expect(await screen.findByText("Editions")).toBeTruthy();
	expect(screen.getByText("Dune (1965)")).toBeTruthy();
});

test("the attribution block is shown only when a TMDB/IGDB field is present", async () => {
	await renderItem(reader(base));
	await screen.findByText("Dune");
	expect(screen.queryByTestId("attribution")).toBeNull();
});

test("a TMDB-sourced field brings the attribution block", async () => {
	await renderItem(reader({ ...base, provenance: { synopsis: { source: "tmdb" } } }));
	expect(await screen.findByTestId("attribution")).toBeTruthy();
});

test("the v1 header shows kicker, score badge with band, verdict and category bars", async () => {
	await renderItem(reader({ ...base, releaseDate: "1965-08-01" }));
	expect(await screen.findByText("Books · 1965")).toBeTruthy();
	expect(screen.getByRole("header", { name: "Dune" })).toBeTruthy();
	expect(screen.getByTestId("score-band-score-high")).toBeTruthy();
	expect(screen.getByText("8.5")).toBeTruthy();
	expect(screen.getByText("Community score · 12 ratings")).toBeTruthy();
	expect(screen.getByText("Category averages")).toBeTruthy();
	expect(screen.getByLabelText("Plot, Score 9.0 out of 10")).toBeTruthy();
	expect(screen.getByRole("button", { name: "Rate it" })).toBeTruthy();
});

test("genre entities become chips and the synopsis text is shown", async () => {
	await renderItem(
		reader({
			...base,
			entities: [{ id: "g1", role: "genre", name: "Science fiction" }],
			synopsis: { text: "Spice and sand." },
		}),
	);
	expect(await screen.findByText("Science fiction")).toBeTruthy();
	expect(screen.getAllByTestId("genre-chip")).toHaveLength(1);
	expect(screen.queryByText("Genres")).toBeNull();
	expect(screen.getByText("Spice and sand.")).toBeTruthy();
});

test("an unscored item shows the not-enough verdict and no category block", async () => {
	await renderItem(
		reader({ ...base, score: null, confidence: null, scoreByCategory: {}, ratingCount: 3 }),
	);
	expect(await screen.findByText("Dune")).toBeTruthy();
	expect(screen.getByTestId("score-band-score-none")).toBeTruthy();
	expect(screen.getByText("Not enough ratings · 3 ratings")).toBeTruthy();
	expect(screen.getByText("No synopsis yet.")).toBeTruthy();
	expect(screen.queryByText("Category averages")).toBeNull();
});

test("the user's own rating from /items/{id}/me is shown", async () => {
	await renderItem(
		reader(base, { data: [] }, { rating: { plot: 7, characters: 8 }, status: null }),
	);
	expect(await screen.findByText("Your overall score")).toBeTruthy();
	expect(screen.getByLabelText("Your overall score, Score 7.5 out of 10")).toBeTruthy();
	expect(screen.getByRole("button", { name: "Update rating" })).toBeTruthy();
});
