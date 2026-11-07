import { fireEvent, render, screen } from "@testing-library/react-native";
import { Text } from "react-native";
import { Avatar, initials } from "./Avatar";
import { Badge } from "./Badge";
import { Button, type ButtonVariant } from "./Button";
import { Card, CardDescription, CardTitle } from "./Card";
import { EmptyState } from "./EmptyState";
import { Input } from "./Input";
import { Label } from "./Label";
import { ListItem } from "./ListItem";
import { SegmentedControl } from "./SegmentedControl";
import { Separator } from "./Separator";
import { Sheet } from "./Sheet";
import { ListSkeleton, Skeleton } from "./Skeleton";
import { TextField } from "./TextField";
import { Toast } from "./Toast";

describe("Button", () => {
	const variants: ButtonVariant[] = ["default", "secondary", "outline", "ghost", "destructive"];

	test.each(variants)("%s is a pressable button with its label", async (variant) => {
		const onPress = jest.fn();
		await render(<Button label="Save" variant={variant} onPress={onPress} />);
		await fireEvent.press(screen.getByRole("button", { name: "Save" }));
		expect(onPress).toHaveBeenCalledTimes(1);
	});

	test("default and outline differ in surface", async () => {
		await render(
			<>
				<Button label="A" onPress={jest.fn()} />
				<Button label="B" variant="outline" onPress={jest.fn()} />
			</>,
		);
		expect(screen.getByRole("button", { name: "A" }).props.className).toContain("bg-primary");
		expect(screen.getByRole("button", { name: "B" }).props.className).toContain("border-border");
	});

	test("sizes keep a touch target: min 44 except small, which pads the hit area", async () => {
		await render(
			<>
				<Button label="S" size="sm" onPress={jest.fn()} />
				<Button label="D" onPress={jest.fn()} />
				<Button label="L" size="lg" onPress={jest.fn()} />
			</>,
		);
		const s = screen.getByRole("button", { name: "S" });
		expect(s.props.hitSlop).toBe(4);
		expect(screen.getByRole("button", { name: "D" }).props.className).toContain("min-h-11");
		expect(screen.getByRole("button", { name: "L" }).props.className).toContain("min-h-12");
	});

	test("disabled and loading do not press and report state", async () => {
		const onPress = jest.fn();
		await render(
			<>
				<Button label="Off" disabled onPress={onPress} />
				<Button label="Busy" loading onPress={onPress} />
			</>,
		);
		await fireEvent.press(screen.getByRole("button", { name: "Off" }));
		await fireEvent.press(screen.getByRole("button", { name: "Busy" }));
		expect(onPress).not.toHaveBeenCalled();
		expect(screen.getByRole("button", { name: "Busy" }).props.accessibilityState).toMatchObject({
			busy: true,
			disabled: true,
		});
	});
});

describe("Input, Label and TextField", () => {
	test("Input is labelled for screen readers and shows the focus ring", async () => {
		await render(<Input label="Email" value="" onChangeText={jest.fn()} />);
		const input = screen.getByLabelText("Email");
		expect(input.props.className).toContain("border-field");
		await fireEvent(input, "focus");
		expect(screen.getByLabelText("Email").props.className).toContain("border-ring");
		await fireEvent(input, "blur");
		expect(screen.getByLabelText("Email").props.className).toContain("border-field");
	});

	test("invalid paints the danger border", async () => {
		await render(<Input label="Email" invalid value="" onChangeText={jest.fn()} />);
		expect(screen.getByLabelText("Email").props.className).toContain("border-danger");
	});

	test("Label renders its text", async () => {
		await render(<Label text="Name" />);
		expect(screen.getByText("Name")).toBeTruthy();
	});

	test("TextField: label, error alert and a password Show / Hide toggle", async () => {
		await render(
			<TextField
				label="Password"
				value="x"
				onChangeText={jest.fn()}
				secure={{ show: "Show", hide: "Hide" }}
				error="Too short"
			/>,
		);
		expect(screen.getByRole("alert")).toHaveTextContent("Too short");
		expect(screen.getByLabelText("Password").props.secureTextEntry).toBe(true);
		await fireEvent.press(screen.getByRole("button", { name: "Show" }));
		expect(screen.getByLabelText("Password").props.secureTextEntry).toBe(false);
		expect(screen.getByRole("button", { name: "Hide" })).toBeTruthy();
	});
});

describe("Card, Badge, Separator, Skeleton", () => {
	test("Card groups a heading and description", async () => {
		await render(
			<Card>
				<CardTitle text="Title" />
				<CardDescription text="Description" />
			</Card>,
		);
		expect(screen.getByRole("header", { name: "Title" })).toBeTruthy();
		expect(screen.getByText("Description")).toBeTruthy();
	});

	test("Badge variants", async () => {
		await render(<Badge label="New" variant="outline" />);
		expect(screen.getByText("New")).toBeTruthy();
	});

	test("Separator and Skeleton are hidden from screen readers", async () => {
		await render(
			<>
				<Separator />
				<Skeleton />
			</>,
		);
		expect(screen.queryByTestId("skeleton")).toBeNull();
		expect(screen.getByTestId("skeleton", { includeHiddenElements: true })).toBeTruthy();
	});
});

describe("Avatar", () => {
	test("initials come from first and last name", () => {
		expect(initials("Ada Lovelace")).toBe("AL");
		expect(initials("  ada  ")).toBe("A");
		expect(initials("")).toBe("?");
	});

	test("shows initials with the name as its label", async () => {
		await render(<Avatar name="Ada Lovelace" />);
		expect(screen.getByLabelText("Ada Lovelace")).toBeTruthy();
		expect(screen.getByText("AL")).toBeTruthy();
	});
});

describe("SegmentedControl", () => {
	test("selects a tab and reports the change", async () => {
		const onChange = jest.fn();
		await render(
			<SegmentedControl
				label="Mode"
				value="a"
				onChange={onChange}
				options={[
					{ value: "a", label: "Alpha" },
					{ value: "b", label: "Beta" },
				]}
			/>,
		);
		expect(screen.getByRole("tab", { name: "Alpha" }).props.accessibilityState.selected).toBe(true);
		await fireEvent.press(screen.getByRole("tab", { name: "Beta" }));
		expect(onChange).toHaveBeenCalledWith("b");
	});
});

describe("Sheet and Toast", () => {
	test("Sheet shows its title and closes from the backdrop", async () => {
		const onClose = jest.fn();
		await render(
			<Sheet visible onClose={onClose} title="Rate" closeLabel="Close">
				<Text>Body</Text>
			</Sheet>,
		);
		expect(screen.getByText("Rate")).toBeTruthy();
		await fireEvent.press(screen.getByRole("button", { name: "Close" }));
		expect(onClose).toHaveBeenCalled();
	});

	test("Sheet renders nothing while hidden", async () => {
		await render(
			<Sheet visible={false} onClose={jest.fn()} title="Rate" closeLabel="Close">
				<Text>Body</Text>
			</Sheet>,
		);
		expect(screen.queryByText("Body")).toBeNull();
	});

	test("Toast is an alert live region", async () => {
		await render(<Toast message="Saved" variant="destructive" />);
		expect(screen.getByText("Saved")).toBeTruthy();
		expect(screen.getByRole("alert").props.accessibilityLiveRegion).toBe("assertive");
	});
});

describe("ListItem and EmptyState", () => {
	test("pressable ListItem is a button named by title and subtitle", async () => {
		const onPress = jest.fn();
		await render(<ListItem title="Zelda" subtitle="Game, 2017" onPress={onPress} />);
		await fireEvent.press(screen.getByRole("button", { name: "Zelda, Game, 2017" }));
		expect(onPress).toHaveBeenCalled();
	});

	test("static ListItem is not a button", async () => {
		await render(<ListItem title="Zelda" />);
		expect(screen.queryByRole("button")).toBeNull();
	});

	test("EmptyState shows a heading, text and an action", async () => {
		const onPress = jest.fn();
		await render(
			<EmptyState title="Nothing" description="Try again" action={{ label: "Browse", onPress }} />,
		);
		expect(screen.getByRole("header", { name: "Nothing" })).toBeTruthy();
		await fireEvent.press(screen.getByRole("button", { name: "Browse" }));
		expect(onPress).toHaveBeenCalled();
	});
});

describe("ListSkeleton", () => {
	test("is one labelled live region of placeholder rows", async () => {
		await render(<ListSkeleton label="Loading" rows={2} />);
		expect(screen.getByLabelText("Loading").props.accessibilityLiveRegion).toBe("polite");
		expect(screen.getAllByTestId("skeleton", { includeHiddenElements: true })).toHaveLength(8);
	});
});
