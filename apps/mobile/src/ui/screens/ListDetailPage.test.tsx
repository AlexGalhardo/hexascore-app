import { fireEvent, render, screen } from "@testing-library/react-native";
import { ListDetailPage, type ListDetailProps } from "./ListDetailPage";

const row = (id: string, title: string) => ({
	id,
	title,
	meta: "Game · 2013",
	coverUri: null,
	score: 8.7,
	scoreLabel: "Score 8.7 out of 10",
});

function setup(overrides: Partial<ListDetailProps> = {}): ListDetailProps {
	return {
		kicker: "Public list · Maya",
		title: "Best of 2013",
		description: "Games that held up.",
		count: "2 titles",
		rows: [row("1", "Grand Theft Auto V"), row("2", "BioShock Infinite")],
		editable: true,
		online: true,
		labels: {
			rename: "Rename",
			remove: "Delete list",
			offline: "Available when you are online",
			moveUp: (r) => `Move ${r.title} up`,
			moveDown: (r) => `Move ${r.title} down`,
			removeItem: (r) => `Remove ${r.title}`,
			empty: "This list is empty.",
		},
		locale: "en",
		onOpen: jest.fn(),
		onRename: jest.fn(),
		onDelete: jest.fn(),
		onMove: jest.fn(),
		onRemoveItem: jest.fn(),
		...overrides,
	};
}

describe("ListDetailPage", () => {
	test("header, rows and owner controls", async () => {
		const props = setup();
		await render(<ListDetailPage {...props} />);
		expect(screen.getByRole("header", { name: "Best of 2013" })).toBeTruthy();
		expect(screen.getByText("2 titles")).toBeTruthy();
		await fireEvent.press(screen.getByRole("button", { name: /BioShock Infinite, Game/ }));
		expect(props.onOpen).toHaveBeenCalledWith(expect.objectContaining({ id: "2" }));
		await fireEvent.press(screen.getByRole("button", { name: "Move BioShock Infinite up" }));
		expect(props.onMove).toHaveBeenCalledWith(expect.objectContaining({ id: "2" }), -1);
		// The first row cannot move up, the last cannot move down.
		expect(
			screen.getByRole("button", { name: "Move Grand Theft Auto V up" }).props.accessibilityState,
		).toMatchObject({ disabled: true });
		expect(
			screen.getByRole("button", { name: "Move BioShock Infinite down" }).props.accessibilityState,
		).toMatchObject({ disabled: true });
		await fireEvent.press(screen.getByRole("button", { name: "Remove Grand Theft Auto V" }));
		expect(props.onRemoveItem).toHaveBeenCalledWith(expect.objectContaining({ id: "1" }));
		await fireEvent.press(screen.getByRole("button", { name: "Rename" }));
		expect(props.onRename).toHaveBeenCalledTimes(1);
	});

	test("offline: rename, delete and reorder are disabled, single remove stays (queued)", async () => {
		await render(<ListDetailPage {...setup({ online: false })} />);
		for (const name of ["Rename", "Delete list", "Move BioShock Infinite up"]) {
			expect(screen.getByRole("button", { name }).props.accessibilityState).toMatchObject({
				disabled: true,
			});
		}
		expect(screen.getByText("Available when you are online")).toBeTruthy();
		expect(
			screen.getByRole("button", { name: "Remove Grand Theft Auto V" }).props.accessibilityState,
		).not.toMatchObject({ disabled: true });
	});

	test("a list that is not mine has no controls; an empty list says so", async () => {
		await render(<ListDetailPage {...setup({ editable: false, rows: [] })} />);
		expect(screen.queryByRole("button", { name: "Rename" })).toBeNull();
		expect(screen.getByText("This list is empty.")).toBeTruthy();
	});
});
