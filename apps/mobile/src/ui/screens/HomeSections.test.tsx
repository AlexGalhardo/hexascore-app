import { fireEvent, render, screen, within } from "@testing-library/react-native";
import { type HomeItem, type HomeLabels, type HomeSection, HomeSections } from "./HomeSections";

const labels: HomeLabels = {
	seeAll: () => "See all",
	rank: (position) => `#${position} trending this week`,
	verdict: (item) => (item.score === null ? "Not enough ratings" : "High score"),
	ratings: (item) =>
		item.score === null ? `${item.ratingCount} of 5 ratings` : `from ${item.ratingCount} ratings`,
	score: (item) => (item.score === null ? "Not enough ratings" : `Score ${item.score} out of 10`),
};

const item = (id: number, title: string, score: number | null = 8.4): HomeItem => ({
	id: String(id),
	title,
	year: 2028,
	coverUri: null,
	score,
	ratingCount: score === null ? 3 : 213,
});

const games: HomeSection = {
	type: "game",
	title: "Games",
	typeLabel: "Game",
	trending: [
		item(1, "The Witcher IV", null),
		item(2, "Diablo V"),
		item(3, "Tomb Raider"),
		item(4, "Extra"),
	],
	grid: Array.from({ length: 8 }, (_, index) => item(10 + index, `Grid game ${index}`, 9)),
};
const books: HomeSection = {
	type: "book",
	title: "Books",
	typeLabel: "Book",
	trending: [],
	grid: [],
};

function setup(sections: HomeSection[] = [games, books]) {
	const onOpen = jest.fn();
	const onSeeAll = jest.fn();
	return { onOpen, onSeeAll, props: { sections, labels, locale: "en", onOpen, onSeeAll } };
}

test("each section has a header and a see-all link", async () => {
	const { props, onSeeAll } = setup();
	await render(<HomeSections {...props} />);
	expect(screen.getAllByRole("header").map((node) => node.props.children)).toEqual([
		"Games",
		"Books",
	]);
	await fireEvent.press(within(screen.getByTestId("home-section-book")).getByRole("link"));
	expect(onSeeAll).toHaveBeenCalledWith("book");
});

test("three trending cards with rank, year · type, verdict and rating count", async () => {
	const { props, onOpen } = setup([games]);
	await render(<HomeSections {...props} />);
	expect(screen.getByText("#1 trending this week")).toBeTruthy();
	expect(screen.getByText("#3 trending this week")).toBeTruthy();
	expect(screen.queryByText("#4 trending this week")).toBeNull();
	expect(screen.queryByText("Extra")).toBeNull();
	expect(screen.getAllByText("2028 · Game")).toHaveLength(3);
	expect(screen.getByText("3 of 5 ratings")).toBeTruthy();
	expect(screen.getAllByText("from 213 ratings")).toHaveLength(2);
	expect(screen.getByText("–")).toBeTruthy();
	await fireEvent.press(screen.getByRole("button", { name: /^#2 trending this week, Diablo V,/ }));
	expect(onOpen).toHaveBeenCalledWith(games.trending[1], "game");
});

test("grid shows six cells with title, year and score", async () => {
	const { props, onOpen } = setup([games]);
	await render(<HomeSections {...props} />);
	expect(screen.getByText("Grid game 5")).toBeTruthy();
	expect(screen.queryByText("Grid game 6")).toBeNull();
	expect(screen.getAllByText("9.0")).toHaveLength(6);
	await fireEvent.press(
		screen.getByRole("button", { name: "Grid game 0, 2028, Score 9 out of 10" }),
	);
	expect(onOpen).toHaveBeenCalledWith(games.grid[0], "game");
});
