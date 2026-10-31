import { fireEvent, render, screen } from "@testing-library/react-native";
import { Cover } from "./Cover";
import { ThemeToggle } from "./ThemeToggle";

test("Cover reserves a 2:3 box and labels the image", async () => {
	await render(<Cover uri="https://cdn.example/c.webp" title="Diablo V" />);
	expect(screen.getByTestId("cover").props.className).toContain("aspect-[2/3]");
	expect(screen.getByLabelText("Diablo V")).toBeTruthy();
});

test("Cover without an image keeps the box and the title for TalkBack", async () => {
	await render(<Cover uri={null} title="Iliad" />);
	expect(screen.getByLabelText("Iliad")).toBeTruthy();
});

test("ThemeToggle shows a moon in light, flips to dark and labels the action", async () => {
	const onChange = jest.fn();
	await render(
		<ThemeToggle
			scheme="light"
			onChange={onChange}
			labels={{ toDark: "Dark theme", toLight: "Light theme" }}
		/>,
	);
	expect(screen.getByText("☾")).toBeTruthy();
	await fireEvent.press(screen.getByRole("button", { name: "Dark theme" }));
	expect(onChange).toHaveBeenCalledWith("dark");
});

test("ThemeToggle shows a sun in dark", async () => {
	const onChange = jest.fn();
	await render(
		<ThemeToggle
			scheme="dark"
			onChange={onChange}
			labels={{ toDark: "Dark theme", toLight: "Light theme" }}
		/>,
	);
	expect(screen.getByText("☀")).toBeTruthy();
	await fireEvent.press(screen.getByRole("button", { name: "Light theme" }));
	expect(onChange).toHaveBeenCalledWith("light");
});
