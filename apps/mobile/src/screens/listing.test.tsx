import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { fireEvent, render, screen, waitFor } from "@testing-library/react-native";
import type {
	ItemJSON,
	ItemPageJSON,
	ListItemsQuery,
} from "../../../../packages/app-client/src/api";
import { ListingBody, type ListingReader } from "./Listing";

function items(prefix: string, count: number): ItemJSON[] {
	return Array.from({ length: count }, (_, i) => ({
		id: `${prefix}${i}`,
		type: "games",
		slug: `${prefix}${i}`,
		locale: "en",
		title: `${prefix} ${i}`,
		score: 7.5,
	}));
}

// Pages keyed by the cursor the client sent ("" = first page).
function reader(pages: Record<string, ItemPageJSON>): ListingReader & { calls: ListItemsQuery[] } {
	const calls: ListItemsQuery[] = [];
	return {
		calls,
		listItems: jest.fn(({ query }) => {
			calls.push(query ?? {});
			return Promise.resolve(pages[query?.cursor ?? ""] ?? { data: [], meta: {} });
		}),
		listGameGenres: jest.fn(() => Promise.resolve({ data: [{ slug: "rpg", count: 3 }] })),
		listMovieGenres: jest.fn(() => Promise.resolve({ data: [] })),
		listSeriesGenres: jest.fn(() => Promise.resolve({ data: [] })),
		listBookGenres: jest.fn(() => Promise.resolve({ data: [] })),
	};
}

async function renderListing(read: ListingReader): Promise<void> {
	const client = new QueryClient({
		defaultOptions: { queries: { retry: false, gcTime: Number.POSITIVE_INFINITY } },
	});
	await render(
		<QueryClientProvider client={client}>
			<ListingBody type="games" read={read} locale="en" />
		</QueryClientProvider>,
	);
}

test("next_cursor drives the next page and both pages render", async () => {
	const read = reader({
		"": { data: items("a", 2), meta: { next_cursor: "c1" } },
		c1: { data: items("b", 1), meta: {} },
	});
	await renderListing(read);
	expect(await screen.findByText("a 0")).toBeTruthy();
	await fireEvent(screen.getByTestId("listing"), "onEndReached");
	expect(await screen.findByText("b 0")).toBeTruthy();
	expect(read.calls.map((q) => q.cursor)).toEqual([undefined, "c1"]);
	expect(read.calls[0]).toMatchObject({ sort: "-popularity", locale: "en" });
});

test("changing the sort or the genre resets the cursor", async () => {
	const read = reader({ "": { data: items("a", 1), meta: { next_cursor: "c1" } } });
	await renderListing(read);
	await screen.findByText("a 0");
	await fireEvent(screen.getByTestId("listing"), "onEndReached");
	await waitFor(() => expect(read.calls).toHaveLength(2));

	await fireEvent.press(screen.getByText("Highest score"));
	await waitFor(() => expect(read.calls).toHaveLength(3));
	expect(read.calls[2]).toEqual({ sort: "-score", locale: "en" });

	await fireEvent.press(await screen.findByText("rpg"));
	await waitFor(() => expect(read.calls).toHaveLength(4));
	expect(read.calls[3]).toEqual({ sort: "-score", genre: "rpg", locale: "en" });
});

test("the end of the list stops fetching", async () => {
	const read = reader({ "": { data: items("a", 2), meta: {} } });
	await renderListing(read);
	await screen.findByText("a 1");
	await fireEvent(screen.getByTestId("listing"), "onEndReached");
	await fireEvent(screen.getByTestId("listing"), "onEndReached");
	expect(read.calls).toHaveLength(1);
});

test("an empty result says so", async () => {
	await renderListing(reader({}));
	expect(await screen.findByText("No items match these filters.")).toBeTruthy();
});
