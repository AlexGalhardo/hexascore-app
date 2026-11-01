import { fireEvent, render, screen } from "@testing-library/react-native";
import type { ItemJSON } from "../../../../packages/app-client/src/api";
import { ItemRow } from "./ItemRow";

const item: ItemJSON = {
	id: "01",
	type: "games",
	slug: "outer-wilds",
	locale: "en",
	title: "Outer Wilds",
	score: 8.5,
	ratingCount: 12,
};

test("a row shows the category label, the title and the score with its band", async () => {
	await render(<ItemRow item={item} locale="en" />);
	expect(screen.getByText("Outer Wilds")).toBeTruthy();
	expect(screen.getByText("Games")).toBeTruthy();
	expect(screen.getByText("8.5")).toBeTruthy();
	expect(screen.getByTestId("score-band-score-high")).toBeTruthy();
	expect(screen.queryByText("Popular")).toBeNull();
});

test("an unscored item still shows a score slot, a dash in the none band", async () => {
	await render(<ItemRow item={{ ...item, score: null }} locale="pt-BR" />);
	expect(screen.getByText("–")).toBeTruthy();
	expect(screen.getByTestId("score-band-score-none")).toBeTruthy();
	expect(screen.getByText("Jogos")).toBeTruthy();
});

test("the score uses the locale's decimal separator", async () => {
	await render(<ItemRow item={{ ...item, score: 7 }} locale="pt-BR" />);
	expect(screen.getByText("7,0")).toBeTruthy();
	expect(screen.getByTestId("score-band-score-mid")).toBeTruthy();
});

test("a prior entry carries the Popular chip", async () => {
	await render(<ItemRow item={item} locale="en" basis="prior" />);
	expect(screen.getByText("Popular")).toBeTruthy();
});

test("the cover thumb renders only when the API sent a URL", async () => {
	await render(<ItemRow item={item} locale="en" />);
	expect(screen.queryByTestId("cover-thumb")).toBeNull();
	await render(<ItemRow item={item} locale="en" coverThumbUrl="https://cdn.test/t.webp" />);
	expect(screen.getByTestId("cover-thumb")).toBeTruthy();
});

test("pressing a row reports its item", async () => {
	const onPress = jest.fn();
	await render(<ItemRow item={item} locale="en" onPress={onPress} />);
	await fireEvent.press(screen.getByText("Outer Wilds"));
	expect(onPress).toHaveBeenCalledWith(item);
});
