import { fireEvent, render, screen } from "@testing-library/react-native";
import {
	type SearchLabels,
	type SearchRow,
	SearchSheet,
	type SearchSheetProps,
} from "./SearchSheet";

const labels: SearchLabels = {
	placeholder: "Search games, movies, series, books…",
	cancel: "Cancel",
	recent: "Recent searches",
	removeRecent: (query) => `Remove ${query}`,
	trending: "Trending this week",
	loading: "Searching…",
	noResults: (query) => `No results for “${query}”`,
	firstOf: (shown, total) => `First ${shown} of ${total} matches`,
	seeAll: (query) => `See all results for “${query}”`,
	score: (row) => (row.score === null ? "Not enough ratings" : `Score ${row.score} out of 10`),
};

const row = (id: number, title: string, score: number | null = 8.7): SearchRow => ({
	id: String(id),
	title,
	year: 2013,
	typeLabel: "Game",
	coverUri: null,
	score,
});

function setup(overrides: Partial<SearchSheetProps>): SearchSheetProps {
	const props: SearchSheetProps = {
		query: "",
		onQueryChange: jest.fn(),
		onCancel: jest.fn(),
		onOpen: jest.fn(),
		onRecent: jest.fn(),
		onRemoveRecent: jest.fn(),
		onSeeAll: jest.fn(),
		state: { kind: "empty", recent: [], trending: [] },
		labels,
		locale: "en",
		...overrides,
	};
	return props;
}

test("focuses the input on open and cancels", async () => {
	const props = setup({});
	await render(<SearchSheet {...props} />);
	const input = screen.getByPlaceholderText(labels.placeholder);
	expect(input.props.autoFocus).toBe(true);
	await fireEvent.changeText(input, "gta");
	expect(props.onQueryChange).toHaveBeenCalledWith("gta");
	await fireEvent.press(screen.getByRole("button", { name: "Cancel" }));
	expect(props.onCancel).toHaveBeenCalled();
});

test("empty state lists recent searches and trending rows", async () => {
	const props = setup({
		state: { kind: "empty", recent: ["witcher"], trending: [row(1, "The Witcher IV", null)] },
	});
	await render(<SearchSheet {...props} />);
	expect(screen.getByText("Recent searches")).toBeTruthy();
	await fireEvent.press(screen.getByRole("button", { name: "witcher" }));
	expect(props.onRecent).toHaveBeenCalledWith("witcher");
	await fireEvent.press(screen.getByRole("button", { name: "Remove witcher" }));
	expect(props.onRemoveRecent).toHaveBeenCalledWith("witcher");
	expect(screen.getByText("Trending this week")).toBeTruthy();
	expect(screen.getByText("–")).toBeTruthy();
});

test("loading and no-results states", async () => {
	const props = setup({ query: "zzz", state: { kind: "loading" } });
	const { rerender } = await render(<SearchSheet {...props} />);
	expect(screen.getByText("Searching…")).toBeTruthy();
	await rerender(<SearchSheet {...props} state={{ kind: "results", rows: [], total: 0 }} />);
	expect(screen.getByText("No results for “zzz”")).toBeTruthy();
	expect(screen.queryByText(/See all/)).toBeNull();
});

test("results show the first 8 rows with year, type and score, and see all", async () => {
	const rows = Array.from({ length: 10 }, (_, index) => row(index, `Grand Theft Auto ${index}`));
	const props = setup({ query: "an", state: { kind: "results", rows, total: 1284 } });
	await render(<SearchSheet {...props} />);
	expect(screen.getByText("First 8 of 1284 matches")).toBeTruthy();
	expect(screen.getAllByText("Game")).toHaveLength(8);
	expect(screen.queryByText("Grand Theft Auto 8")).toBeNull();
	expect(screen.getAllByText("2013")).toHaveLength(8);
	expect(screen.getAllByText("8.7")).toHaveLength(8);
	await fireEvent.press(screen.getByRole("button", { name: /Grand Theft Auto 1,/ }));
	expect(props.onOpen).toHaveBeenCalledWith(rows[1]);
	await fireEvent.press(screen.getByRole("button", { name: "See all results for “an”" }));
	expect(props.onSeeAll).toHaveBeenCalled();
});
