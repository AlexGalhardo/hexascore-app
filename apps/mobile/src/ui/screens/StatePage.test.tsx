import { fireEvent, render, screen } from "@testing-library/react-native";
import { StatePage } from "./StatePage";

describe("StatePage", () => {
	test("404: code, title as header, body and actions in order", async () => {
		const home = jest.fn();
		const search = jest.fn();
		await render(
			<StatePage
				code="404"
				title="Page not found"
				body={["The page you are looking for does not exist or was moved."]}
				actions={[
					{ label: "Back to home", onPress: home },
					{ label: "Search", onPress: search, variant: "secondary" },
				]}
			/>,
		);
		expect(screen.getByText("404", { includeHiddenElements: true })).toBeTruthy();
		expect(screen.getByRole("header", { name: "Page not found" })).toBeTruthy();
		expect(screen.getByText(/does not exist/)).toBeTruthy();
		const buttons = screen.getAllByRole("button").map((b) => b.props.accessibilityLabel);
		expect(buttons).toEqual(["Back to home", "Search"]);
		await fireEvent.press(screen.getByRole("button", { name: "Search" }));
		expect(search).toHaveBeenCalledTimes(1);
		expect(home).not.toHaveBeenCalled();
	});

	test("without a code the big number is not drawn and the region is announced", async () => {
		await render(<StatePage title="No public items yet." body={[]} actions={[]} />);
		expect(screen.queryByText("404")).toBeNull();
		expect(screen.getByTestId("state-page").props.accessibilityLiveRegion).toBe("polite");
	});

	test("the code is decorative: hidden from accessibility so the title is read once", async () => {
		await render(<StatePage code="500" title="Something went wrong" body={[]} actions={[]} />);
		expect(screen.queryByText("500")).toBeNull();
		expect(screen.getByText("500", { includeHiddenElements: true })).toBeTruthy();
	});

	test("a link action renders with the link role", async () => {
		const open = jest.fn();
		await render(
			<StatePage
				title="Update required"
				body={[]}
				actions={[{ label: "Open the store", onPress: open, role: "link" }]}
			/>,
		);
		await fireEvent.press(screen.getByRole("link", { name: "Open the store" }));
		expect(open).toHaveBeenCalledTimes(1);
	});
});
