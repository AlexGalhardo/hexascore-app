import { fireEvent, render, screen } from "@testing-library/react-native";
import { palette } from "../theme";
import { AppHeader } from "./AppHeader";
import { tabBarTheme } from "./tabBarTheme";

const labels = { search: "Open search", toDark: "Dark theme", toLight: "Light theme" };

describe("AppHeader", () => {
	test("shows the serif wordmark, opens search and toggles the theme", async () => {
		const onSearch = jest.fn();
		const onToggleTheme = jest.fn();
		await render(
			<AppHeader
				scheme="light"
				topInset={24}
				labels={labels}
				onSearch={onSearch}
				onToggleTheme={onToggleTheme}
			/>,
		);
		const wordmark = screen.getByRole("header");
		expect(wordmark.props.children).toBe("HexaScore");
		expect(wordmark.props.className).toContain("font-display");
		await fireEvent.press(screen.getByRole("button", { name: "Open search" }));
		expect(onSearch).toHaveBeenCalledTimes(1);
		await fireEvent.press(screen.getByRole("button", { name: "Dark theme" }));
		expect(onToggleTheme).toHaveBeenCalledWith("dark");
	});

	test("keeps clear of the status bar", async () => {
		await render(
			<AppHeader
				scheme="dark"
				topInset={32}
				labels={labels}
				onSearch={jest.fn()}
				onToggleTheme={jest.fn()}
			/>,
		);
		expect(screen.getByTestId("app-header").props.style).toEqual({ paddingTop: 32 });
		expect(screen.getByRole("button", { name: "Light theme" })).toBeTruthy();
	});
});

describe("tabBarTheme", () => {
	test("active tab uses primary, the rest muted, on the page background with a hairline", () => {
		for (const scheme of ["light", "dark"] as const) {
			const colors = palette(scheme);
			expect(tabBarTheme(scheme)).toEqual({
				tabBarActiveTintColor: colors.primary,
				tabBarInactiveTintColor: colors.muted,
				tabBarStyle: { backgroundColor: colors.bg, borderTopColor: colors.border },
				tabBarLabelStyle: { fontFamily: "Inter", fontWeight: "600", fontSize: 12 },
			});
		}
	});
});
