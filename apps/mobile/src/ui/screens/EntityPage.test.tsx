import { fireEvent, render, screen } from "@testing-library/react-native";
import { EntityPage, type EntityPageProps } from "./EntityPage";

const row = (id: string, title: string, score: number | null) => ({
	id,
	title,
	meta: "Game · 1993",
	coverUri: null,
	score,
	scoreLabel: score === null ? "Not enough ratings" : `Score ${score} out of 10`,
});

function setup(overrides: Partial<EntityPageProps> = {}): EntityPageProps {
	return {
		kind: "Company",
		name: "Blizzard Entertainment",
		roles: [{ key: "developer", label: "Developer", count: 2 }],
		activeRole: "developer",
		facts: [{ label: "Founded", value: "1991" }],
		description: "American video game developer and publisher.",
		licence: "Description licence: CC BY-SA 4.0",
		sourceLabel: "Source",
		groups: [
			{
				key: "game",
				title: "Games",
				count: "2 of 2",
				note: "Average score: 7.8 (items with a score only)",
				rows: [row("1", "The Lost Vikings", 7.8), row("2", "Diablo V", null)],
			},
		],
		empty: {
			title: "No public items yet.",
			body: "Items will appear here.",
			action: "Browse games",
		},
		locale: "en",
		onRole: jest.fn(),
		onOpen: jest.fn(),
		onSource: jest.fn(),
		onBrowse: jest.fn(),
		...overrides,
	};
}

describe("EntityPage", () => {
	test("header, role filter, facts, description and grouped rows", async () => {
		const props = setup();
		await render(<EntityPage {...props} />);
		expect(screen.getByText("Company")).toBeTruthy();
		expect(screen.getByRole("header", { name: "Blizzard Entertainment" })).toBeTruthy();
		const role = screen.getByRole("button", { name: "Developer (2)" });
		expect(role.props.accessibilityState).toMatchObject({ selected: true });
		await fireEvent.press(role);
		expect(props.onRole).toHaveBeenCalledWith("developer");
		expect(screen.getByText("Founded")).toBeTruthy();
		expect(screen.getByText("1991")).toBeTruthy();
		expect(screen.getByRole("header", { name: "Games" })).toBeTruthy();
		expect(screen.getByText("2 of 2")).toBeTruthy();
		await fireEvent.press(screen.getByRole("button", { name: /Diablo V/ }));
		expect(props.onOpen).toHaveBeenCalledWith(expect.objectContaining({ id: "2" }));
		await fireEvent.press(screen.getByRole("link", { name: "Source" }));
		expect(props.onSource).toHaveBeenCalledTimes(1);
	});

	test("groups with no rows are hidden", async () => {
		await render(
			<EntityPage
				{...setup({
					groups: [
						{ key: "game", title: "Games", count: "1", rows: [row("1", "Diablo V", null)] },
						{ key: "book", title: "Books", count: "0", rows: [] },
					],
				})}
			/>,
		);
		expect(screen.queryByText("Books")).toBeNull();
	});

	test("no items at all shows the empty state with a browse action", async () => {
		const props = setup({ groups: [], roles: [], facts: [], description: null });
		await render(<EntityPage {...props} />);
		expect(screen.getByText("No public items yet.")).toBeTruthy();
		expect(screen.queryByRole("link", { name: "Source" })).toBeNull();
		await fireEvent.press(screen.getByRole("button", { name: "Browse games" }));
		expect(props.onBrowse).toHaveBeenCalledTimes(1);
	});
});
