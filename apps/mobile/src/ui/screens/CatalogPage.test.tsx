import { fireEvent, render, screen } from "@testing-library/react-native";
import { Text } from "react-native";
import {
	type CatalogLabels,
	CatalogPage,
	type CatalogPageProps,
	type CatalogRow,
} from "./CatalogPage";

const labels: CatalogLabels = {
	filters: (active) => (active > 0 ? `Filters · ${active}` : "Filters"),
	closeFilters: "Close filters",
	sort: (current) => `Sort: ${current}`,
	found: (total) => `${total} items found`,
	previous: "Previous",
	next: "Next",
	page: (page, pages) => `Page ${page} of ${pages}`,
	apply: "Apply filters",
	reset: "Reset",
	empty: "No items match these filters.",
	meta: (row) => [row.year, row.byline, `${row.ratingCount} ratings`].filter(Boolean).join(" · "),
	score: (row) => (row.score === null ? "Not enough ratings" : `Score ${row.score} out of 10`),
};

const row = (id: number, title: string, score: number | null = 9): CatalogRow => ({
	id: String(id),
	title,
	year: 2009,
	byline: "Mojang Studios",
	ratingCount: 1284,
	coverUri: null,
	score,
});

function setup(overrides: Partial<CatalogPageProps> = {}): CatalogPageProps {
	return {
		title: "Games",
		rows: [row(1, "Minecraft"), row(2, "Roblox", 5.4)],
		total: 24873,
		page: 1,
		pages: 1244,
		activeFilters: 1,
		sortLabel: "Most rated",
		filtersOpen: false,
		filterPanel: <Text>genre field</Text>,
		labels,
		locale: "en",
		onToggleFilters: jest.fn(),
		onSort: jest.fn(),
		onApply: jest.fn(),
		onReset: jest.fn(),
		onPage: jest.fn(),
		onOpen: jest.fn(),
		...overrides,
	};
}

test("title, count and rows with meta and score", async () => {
	const props = setup();
	await render(<CatalogPage {...props} />);
	expect(screen.getByRole("header").props.children).toBe("Games");
	expect(screen.getByText("24873 items found")).toBeTruthy();
	expect(screen.getAllByText("2009 · Mojang Studios · 1284 ratings")).toHaveLength(2);
	expect(screen.getByText("5.4")).toBeTruthy();
	await fireEvent.press(screen.getByRole("button", { name: /^Minecraft,/ }));
	expect(props.onOpen).toHaveBeenCalledWith(props.rows[0]);
});

test("pager top and bottom: previous disabled on page 1, next goes forward", async () => {
	const props = setup();
	await render(<CatalogPage {...props} />);
	expect(screen.getAllByText("Page 1 of 1244")).toHaveLength(2);
	const [previous] = screen.getAllByRole("button", { name: "Previous" });
	expect(previous.props.accessibilityState).toMatchObject({ disabled: true });
	await fireEvent.press(screen.getAllByRole("button", { name: "Next" })[1]);
	expect(props.onPage).toHaveBeenCalledWith(2);
});

test("next disabled on the last page", async () => {
	const props = setup({ page: 1244 });
	await render(<CatalogPage {...props} />);
	const [next] = screen.getAllByRole("button", { name: "Next" });
	expect(next.props.accessibilityState).toMatchObject({ disabled: true });
	await fireEvent.press(screen.getAllByRole("button", { name: "Previous" })[0]);
	expect(props.onPage).toHaveBeenCalledWith(1243);
});

test("filters toggle, sort trigger, open panel with apply and reset", async () => {
	const props = setup();
	const { rerender } = await render(<CatalogPage {...props} />);
	expect(screen.queryByText("genre field")).toBeNull();
	await fireEvent.press(screen.getByRole("button", { name: "Filters · 1" }));
	expect(props.onToggleFilters).toHaveBeenCalled();
	await fireEvent.press(screen.getByRole("button", { name: "Sort: Most rated" }));
	expect(props.onSort).toHaveBeenCalled();

	await rerender(<CatalogPage {...props} filtersOpen />);
	expect(screen.getByText("genre field")).toBeTruthy();
	expect(
		screen.getByRole("button", { name: "Close filters" }).props.accessibilityState,
	).toMatchObject({
		expanded: true,
	});
	expect(screen.queryByText("Minecraft")).toBeNull();
	await fireEvent.press(screen.getByRole("button", { name: "Apply filters" }));
	expect(props.onApply).toHaveBeenCalled();
	await fireEvent.press(screen.getByRole("button", { name: "Reset" }));
	expect(props.onReset).toHaveBeenCalled();
});

test("empty result shows the message and reset, no pager", async () => {
	const props = setup({ rows: [], total: 0, pages: 0 });
	await render(<CatalogPage {...props} />);
	expect(screen.getByText("No items match these filters.")).toBeTruthy();
	expect(screen.queryByText(/Page/)).toBeNull();
	await fireEvent.press(screen.getByRole("button", { name: "Reset" }));
	expect(props.onReset).toHaveBeenCalled();
});
