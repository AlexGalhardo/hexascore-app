import { fireEvent, render, screen } from "@testing-library/react-native";
import { TextField } from "../TextField";
import { AuthPage, type AuthPageProps } from "./AuthPage";

function setup(overrides: Partial<AuthPageProps> = {}): AuthPageProps {
	return {
		kicker: "Members",
		title: "Log in to rate, shelve and list.",
		lede: "One account for the site and the Android app.",
		google: { label: "Sign in with Google", scheme: "light", onPress: jest.fn() },
		orLabel: "or",
		footer: { prompt: "New here?", action: "Create an account", onPress: jest.fn() },
		children: null,
		...overrides,
	};
}

test("header, Google button above the or divider, footer link", async () => {
	const props = setup();
	await render(<AuthPage {...props} />);
	expect(screen.getByText("Members")).toBeOnTheScreen();
	expect(
		screen.getByRole("header", { name: "Log in to rate, shelve and list." }),
	).toBeOnTheScreen();
	expect(screen.getByText("One account for the site and the Android app.")).toBeOnTheScreen();
	expect(screen.getByTestId("google-mark")).toBeOnTheScreen();
	// The divider is decorative, so it is hidden from screen readers.
	expect(screen.getByText("or", { includeHiddenElements: true })).toBeOnTheScreen();
	await fireEvent.press(screen.getByRole("button", { name: "Sign in with Google" }));
	expect(props.google.onPress).toHaveBeenCalled();
	await fireEvent.press(screen.getByRole("link", { name: "Create an account" }));
	expect(props.footer.onPress).toHaveBeenCalled();
});

test("Google button follows the brand colours per scheme and can be disabled", async () => {
	await render(
		<AuthPage
			{...setup({
				google: {
					label: "Sign in with Google",
					scheme: "dark",
					onPress: jest.fn(),
					disabled: true,
				},
			})}
		/>,
	);
	const button = screen.getByRole("button", { name: "Sign in with Google" });
	expect(button).toBeDisabled();
	expect(button.props.className ?? "").toContain("bg-[#131314]");
});

test("form slot renders between the divider and the footer", async () => {
	await render(
		<AuthPage {...setup()}>
			<TextField label="Email or nickname" value="maya" onChangeText={jest.fn()} />
		</AuthPage>,
	);
	expect(screen.getByLabelText("Email or nickname")).toBeOnTheScreen();
});

test("password field toggles visibility with a labelled control", async () => {
	await render(
		<TextField
			label="Password"
			value="secret-password"
			onChangeText={jest.fn()}
			secure={{ show: "Show", hide: "Hide" }}
		/>,
	);
	const input = screen.getByLabelText("Password");
	expect(input.props.secureTextEntry).toBe(true);
	await fireEvent.press(screen.getByRole("button", { name: "Show" }));
	expect(screen.getByLabelText("Password").props.secureTextEntry).toBe(false);
	expect(screen.getByRole("button", { name: "Hide" })).toBeOnTheScreen();
});

test("field error is announced", async () => {
	await render(
		<TextField label="Email" value="" onChangeText={jest.fn()} error="Enter your email." />,
	);
	expect(screen.getByRole("alert")).toHaveTextContent("Enter your email.");
});
