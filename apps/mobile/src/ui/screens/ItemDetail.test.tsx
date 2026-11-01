import { fireEvent, render, screen } from "@testing-library/react-native";
import {
	ItemDetail,
	type ItemDetailLabels,
	type ItemDetailProps,
	type ItemDetailView,
} from "./ItemDetail";

const labels: ItemDetailLabels = {
	verdict: (score) => (score === null ? "Not enough ratings" : "High"),
	ratings: (count) => `${count} ratings`,
	score: (score) => (score === null ? "Not enough ratings" : `Score ${score} out of 10`),
	rate: "Rate",
	addToList: "Add to list",
	categoryAverages: "Category averages",
	synopsis: "Synopsis",
	noSynopsis: "No synopsis yet.",
	rateTitle: (title) => `Rate ${title}`,
	overall: "Your overall score",
	save: "Save rating",
	moreLikeThis: "More like this",
};

const item: ItemDetailView = {
	title: "Diablo V",
	kicker: "Game · 2029",
	byline: "Blizzard Entertainment",
	coverUri: null,
	score: 8.4,
	ratingCount: 213,
	genres: ["Action RPG", "Dark fantasy"],
	synopsis: "Lilith returns.",
	categories: [
		{ key: "gameplay", label: "Gameplay", score: 8.8 },
		{ key: "technical", label: "Technical", score: null },
	],
};

function setup(overrides: Partial<ItemDetailProps> = {}): ItemDetailProps {
	return {
		item,
		statusLabel: "Want to play",
		similar: [{ id: "7", title: "The Witcher IV", year: 2028, coverUri: null, score: null }],
		rating: null,
		labels,
		locale: "en",
		onRate: jest.fn(),
		onStatus: jest.fn(),
		onAddToList: jest.fn(),
		onRateCategory: jest.fn(),
		onSave: jest.fn(),
		onOpenSimilar: jest.fn(),
		...overrides,
	};
}

test("header: kicker, title, byline, 44 px genre chips and score block", async () => {
	await render(<ItemDetail {...setup()} />);
	expect(screen.getByText("Game · 2029")).toBeTruthy();
	expect(screen.getByRole("header", { name: "Diablo V" })).toBeTruthy();
	expect(screen.getByText("Blizzard Entertainment")).toBeTruthy();
	expect(screen.getByText("Action RPG")).toBeTruthy();
	for (const chip of screen.getAllByTestId("genre-chip"))
		expect(chip.props.className).toContain("min-h-11");
	expect(screen.getByLabelText("Score 8.4 out of 10")).toBeTruthy();
	expect(screen.getByText("High · 213 ratings")).toBeTruthy();
});

test("actions call back", async () => {
	const props = setup();
	await render(<ItemDetail {...props} />);
	await fireEvent.press(screen.getByRole("button", { name: "Rate" }));
	await fireEvent.press(screen.getByRole("button", { name: "Want to play" }));
	await fireEvent.press(screen.getByRole("button", { name: "Add to list" }));
	expect(props.onRate).toHaveBeenCalled();
	expect(props.onStatus).toHaveBeenCalled();
	expect(props.onAddToList).toHaveBeenCalled();
});

test("category averages print the number, '–' without one", async () => {
	await render(<ItemDetail {...setup()} />);
	expect(screen.getByText("Category averages")).toBeTruthy();
	expect(screen.getByLabelText("Gameplay, Score 8.8 out of 10")).toBeTruthy();
	expect(screen.getByLabelText("Technical, Not enough ratings")).toBeTruthy();
	expect(screen.getByText("8.8")).toBeTruthy();
});

test("rating panel: one adjustable input per category, overall and save", async () => {
	const props = setup({
		rating: { values: { gameplay: 7.5, technical: null }, overall: 7.5, saving: false },
	});
	await render(<ItemDetail {...props} />);
	expect(screen.getByText("Rate Diablo V")).toBeTruthy();
	const inputs = screen.getAllByRole("adjustable");
	expect(inputs).toHaveLength(2);
	await fireEvent(inputs[0], "accessibilityAction", { nativeEvent: { actionName: "increment" } });
	expect(props.onRateCategory).toHaveBeenCalledWith("gameplay", 8);
	expect(screen.getByLabelText("Your overall score, Score 7.5 out of 10")).toBeTruthy();
	await fireEvent.press(screen.getByRole("button", { name: "Save rating" }));
	expect(props.onSave).toHaveBeenCalled();
});

test("saving disables the save button", async () => {
	await render(<ItemDetail {...setup({ rating: { values: {}, overall: null, saving: true } })} />);
	expect(
		screen.getByRole("button", { name: "Save rating" }).props.accessibilityState,
	).toMatchObject({
		disabled: true,
	});
});

test("no score, no synopsis and more like this", async () => {
	const props = setup({ item: { ...item, score: null, ratingCount: 3, synopsis: null } });
	await render(<ItemDetail {...props} />);
	expect(screen.getByText("Not enough ratings · 3 ratings")).toBeTruthy();
	expect(screen.getByText("No synopsis yet.")).toBeTruthy();
	expect(screen.getByText("More like this")).toBeTruthy();
	await fireEvent.press(screen.getByRole("button", { name: /^The Witcher IV,/ }));
	expect(props.onOpenSimilar).toHaveBeenCalledWith(props.similar[0]);
});
