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
		slug: `${prefix}-slug-${i}`,
		locale: "en",
		title: `${prefix} ${i}`,
		score: 7.5,
		ratingCount: 12,
		releaseDate: "2009-05-17",
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

async function renderListing(read: ListingReader): Promise<jest.Mock> {
	const client = new QueryClient({
		defaultOptions: {
			queries: { retry: false, gcTime: Number.POSITIVE_INFINITY, staleTime: 30_000 },
		},
	});
	const onOpen = jest.fn();
	await render(
		<QueryClientProvider client={client}>
			<ListingBody type="games" read={read} locale="en" onOpen={onOpen} />
		</QueryClientProvider>,
	);
	return onOpen;
}

test("the first page shows title, rows with year · rating count and score", async () => {
	const read = reader({ "": { data: items("a", 2), meta: {} } });
	const onOpen = await renderListing(read);
	expect(await screen.findByText("a 0")).toBeTruthy();
	expect(screen.getByRole("header").props.children).toBe("Games");
	expect(screen.getAllByText("2009 · 12 ratings")).toHaveLength(2);
	// The cursor API sends no total, so no "N items found" line.
	expect(screen.queryByText(/found/)).toBeNull();
	expect(read.calls[0]).toEqual({ sort: "-popularity", locale: "en" });
	await fireEvent.press(screen.getByRole("button", { name: /^a 1,/ }));
	expect(onOpen).toHaveBeenCalledWith(expect.objectContaining({ id: "a1", slug: "a-slug-1" }));
});

test("Next follows next_cursor, Previous goes back from the cache", async () => {
	const read = reader({
		"": { data: items("a", 2), meta: { next_cursor: "c1" } },
		c1: { data: items("b", 1), meta: {} },
	});
	await renderListing(read);
	await screen.findByText("a 0");
	expect(screen.getAllByText("Page 1")).toHaveLength(2);

	await fireEvent.press(screen.getAllByRole("button", { name: "Next" })[0]);
	expect(await screen.findByText("b 0")).toBeTruthy();
	expect(screen.queryByText("a 0")).toBeNull();
	expect(screen.getAllByText("Page 2 of 2")).toHaveLength(2);
	expect(screen.getAllByRole("button", { name: "Next" })[0].props.accessibilityState).toMatchObject(
		{ disabled: true },
	);
	expect(read.calls.map((q) => q.cursor)).toEqual([undefined, "c1"]);

	await fireEvent.press(screen.getAllByRole("button", { name: "Previous" })[0]);
	expect(await screen.findByText("a 0")).toBeTruthy();
	expect(read.calls).toHaveLength(2);
});

test("sort and genre are applied from the filter panel and reset the cursor", async () => {
	const read = reader({ "": { data: items("a", 1), meta: { next_cursor: "c1" } } });
	await renderListing(read);
	await screen.findByText("a 0");
	await fireEvent.press(screen.getAllByRole("button", { name: "Next" })[0]);
	await waitFor(() => expect(read.calls).toHaveLength(2));

	await fireEvent.press(screen.getByRole("button", { name: "Filters" }));
	await fireEvent.press(screen.getByRole("button", { name: "Highest score" }));
	await fireEvent.press(await screen.findByRole("button", { name: "rpg" }));
	await fireEvent.press(screen.getByRole("button", { name: "Apply" }));
	await waitFor(() => expect(read.calls).toHaveLength(3));
	expect(read.calls[2]).toEqual({ sort: "-score", genre: "rpg", locale: "en" });
	expect(await screen.findByRole("button", { name: "Filters · 1" })).toBeTruthy();
	expect(screen.getByRole("button", { name: "Sort by: Highest score" })).toBeTruthy();
	expect(screen.getAllByText("Page 1")).toHaveLength(2);
});

test("the sort trigger opens the panel and Reset restores the defaults", async () => {
	const read = reader({ "": { data: items("a", 1), meta: {} } });
	await renderListing(read);
	await screen.findByText("a 0");
	await fireEvent.press(screen.getByRole("button", { name: "Sort by: Most popular" }));
	await fireEvent.press(await screen.findByRole("button", { name: "rpg" }));
	await fireEvent.press(screen.getByRole("button", { name: "Apply" }));
	await waitFor(() => expect(read.calls).toHaveLength(2));

	await fireEvent.press(await screen.findByRole("button", { name: "Filters · 1" }));
	await fireEvent.press(screen.getByRole("button", { name: "Reset" }));
	expect(await screen.findByRole("button", { name: "Filters" })).toBeTruthy();
	// The default filter page is still fresh (app staleTime 30 s), so Reset does not refetch.
	expect(read.calls).toHaveLength(2);
	expect(screen.getByRole("button", { name: "Sort by: Most popular" })).toBeTruthy();
});

test("an empty result says so", async () => {
	await renderListing(reader({}));
	expect(await screen.findByText("No items match these filters.")).toBeTruthy();
});
