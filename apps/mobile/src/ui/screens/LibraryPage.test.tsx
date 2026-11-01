import { fireEvent, render, screen } from "@testing-library/react-native";
import {
	type LibraryGroup,
	type LibraryLabels,
	LibraryPage,
	type LibraryPageProps,
	type LibraryRow,
} from "./LibraryPage";

const labels: LibraryLabels = {
	summary: (titles, rated) => `${titles} titles · ${rated} rated`,
	statusField: "Status",
	status: (current) => `Status: ${current}`,
	filters: "Filters",
	groupCount: (group) =>
		group.rows.length < group.total
			? `${group.total} titles · showing ${group.rows.length}`
			: `${group.total} titles`,
	seeAll: (group) => `See all ${group.total}`,
	meta: (row) =>
		[row.typeLabel, row.year, row.myScore === null ? null : `your score ${row.myScore}`]
			.filter(Boolean)
			.join(" · "),
	score: (row) => (row.score === null ? "Not enough ratings" : `Score ${row.score} out of 10`),
	empty: "Nothing on your shelves yet.",
	browse: "Browse the catalogue",
};

const row = (
	id: number,
	title: string,
	score: number | null,
	myScore: number | null,
): LibraryRow => ({
	id: String(id),
	title,
	typeLabel: "Game",
	year: 2013,
	coverUri: null,
	score,
	myScore,
});

const groups: LibraryGroup[] = [
	{
		key: "in_progress",
		title: "In progress",
		total: 2,
		rows: [row(1, "The Witcher IV", null, null)],
	},
	{ key: "completed", title: "Completed", total: 5, rows: [row(2, "Grand Theft Auto V", 8.7, 9)] },
];

function setup(overrides: Partial<LibraryPageProps> = {}): LibraryPageProps {
	return {
		kicker: "Maya's shelves",
		title: "My library",
		titles: 14,
		rated: 9,
		statusLabel: "All statuses (14)",
		groups,
		labels,
		locale: "en",
		onStatus: jest.fn(),
		onFilters: jest.fn(),
		onOpen: jest.fn(),
		onSeeAll: jest.fn(),
		onBrowse: jest.fn(),
		...overrides,
	};
}

test("header, summary, status trigger and filters", async () => {
	const props = setup();
	await render(<LibraryPage {...props} />);
	expect(screen.getByText("Maya's shelves")).toBeOnTheScreen();
	expect(screen.getByRole("header", { name: "My library" })).toBeOnTheScreen();
	expect(screen.getByText("14 titles · 9 rated")).toBeOnTheScreen();
	await fireEvent.press(screen.getByRole("button", { name: "Status: All statuses (14)" }));
	expect(props.onStatus).toHaveBeenCalled();
	await fireEvent.press(screen.getByRole("button", { name: "Filters" }));
	expect(props.onFilters).toHaveBeenCalled();
});

test("groups with count, rows with meta and score, see all only when truncated", async () => {
	const props = setup();
	await render(<LibraryPage {...props} />);
	expect(screen.getByRole("header", { name: "In progress" })).toBeOnTheScreen();
	expect(screen.getByText("2 titles · showing 1")).toBeOnTheScreen();
	expect(screen.getByText("5 titles · showing 1")).toBeOnTheScreen();
	expect(screen.getByText("Game · 2013 · your score 9")).toBeOnTheScreen();
	expect(screen.getByText("8.7")).toBeOnTheScreen();
	await fireEvent.press(
		screen.getByRole("button", {
			name: "Grand Theft Auto V, Game · 2013 · your score 9, Score 8.7 out of 10",
		}),
	);
	expect(props.onOpen).toHaveBeenCalledWith(groups[1]?.rows[0]);
	await fireEvent.press(screen.getByRole("link", { name: "See all 5" }));
	expect(props.onSeeAll).toHaveBeenCalledWith(groups[1]);
});

test("a complete group has no see all link", async () => {
	const full: LibraryGroup = { ...(groups[1] as LibraryGroup), total: 1 };
	await render(<LibraryPage {...setup({ groups: [full] })} />);
	expect(screen.getByText("1 titles")).toBeOnTheScreen();
	expect(screen.queryByRole("link")).toBeNull();
});

test("empty library offers a way to the catalogue", async () => {
	const props = setup({ groups: [], titles: 0, rated: 0 });
	await render(<LibraryPage {...props} />);
	expect(screen.getByText("Nothing on your shelves yet.")).toBeOnTheScreen();
	await fireEvent.press(screen.getByRole("button", { name: "Browse the catalogue" }));
	expect(props.onBrowse).toHaveBeenCalled();
});
