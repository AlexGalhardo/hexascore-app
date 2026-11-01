import { fireEvent, render, screen } from "@testing-library/react-native";
import { Text } from "react-native";
import { type SettingsLabels, SettingsPage, type SettingsPageProps } from "./SettingsPage";

const labels: SettingsLabels = {
	account: "Account",
	appearance: "Appearance",
	theme: { system: "System", light: "Light", dark: "Dark" },
	privacy: "Privacy",
	signOut: "Log out",
	signIn: "Log in",
};

function setup(overrides: Partial<SettingsPageProps> = {}): SettingsPageProps {
	return {
		title: "Settings",
		account: { name: "maya", email: "maya@example.com" },
		theme: "system",
		crashReport: <Text>crash toggle</Text>,
		labels,
		onTheme: jest.fn(),
		onSignOut: jest.fn(),
		onSignIn: jest.fn(),
		...overrides,
	};
}

test("account block, theme mode radios and crash report slot", async () => {
	const props = setup();
	await render(<SettingsPage {...props} />);
	expect(screen.getByRole("header", { name: "Settings" })).toBeOnTheScreen();
	expect(screen.getByText("maya")).toBeOnTheScreen();
	expect(screen.getByText("maya@example.com")).toBeOnTheScreen();
	expect(screen.getByRole("radio", { name: "System" })).toBeChecked();
	expect(screen.getByRole("radio", { name: "Dark" })).not.toBeChecked();
	await fireEvent.press(screen.getByRole("radio", { name: "Dark" }));
	expect(props.onTheme).toHaveBeenCalledWith("dark");
	expect(screen.getByRole("header", { name: "Privacy" })).toBeOnTheScreen();
	expect(screen.getByText("crash toggle")).toBeOnTheScreen();
	await fireEvent.press(screen.getByRole("button", { name: "Log out" }));
	expect(props.onSignOut).toHaveBeenCalled();
});

test("signed out shows log in instead of the account", async () => {
	const props = setup({ account: null });
	await render(<SettingsPage {...props} />);
	expect(screen.queryByText("maya@example.com")).toBeNull();
	await fireEvent.press(screen.getByRole("button", { name: "Log in" }));
	expect(props.onSignIn).toHaveBeenCalled();
});
