import { fireEvent, render, screen } from "@testing-library/react-native";
import type { ReactElement, ReactNode } from "react";
import TabsLayout from "../app/(tabs)/_layout";
import { LocaleProvider } from "./providers/root";

type Options = Record<string, unknown> & {
	title?: string;
	header?: () => ReactElement;
	tabBarIcon?: (props: { focused: boolean; color: string; size: number }) => ReactElement;
};
const mockCaptured: { screenOptions?: Options; screens: { name: string; options: Options }[] } = {
	screens: [],
};
const mockNavigate = jest.fn();

jest.mock("expo-router", () => {
	function Tabs(props: { screenOptions: Options; children?: ReactNode }): ReactNode {
		// A re-render (colour scheme subscription) renders the screens again; keep the last pass.
		mockCaptured.screenOptions = props.screenOptions;
		mockCaptured.screens = [];
		return props.children ?? null;
	}
	Tabs.Screen = function Screen(props: { name: string; options: Options }): null {
		mockCaptured.screens.push({ name: props.name, options: props.options });
		return null;
	};
	return { Tabs, useRouter: () => ({ navigate: mockNavigate }) };
});

jest.mock("react-native-safe-area-context", () => ({
	useSafeAreaInsets: () => ({ top: 24, bottom: 0, left: 0, right: 0 }),
}));

beforeEach(() => {
	mockCaptured.screens = [];
	mockNavigate.mockClear();
});

test("tabs in round 2 order with translated titles and icons", async () => {
	await render(
		<LocaleProvider locale="pt-BR">
			<TabsLayout />
		</LocaleProvider>,
	);
	expect(mockCaptured.screens.map((s) => s.name)).toEqual([
		"index",
		"search",
		"[type]",
		"lists",
		"profile",
	]);
	expect(mockCaptured.screens.map((s) => s.options.title)).toEqual([
		"Início",
		"Buscar",
		"Seções",
		"Minhas listas",
		"Perfil",
	]);
	for (const tab of mockCaptured.screens) {
		expect(tab.options.tabBarIcon?.({ focused: true, color: "#000", size: 24 })).toBeTruthy();
	}
	expect(mockCaptured.screenOptions?.tabBarActiveTintColor).toEqual(expect.any(String));
});

test("header search opens the search tab and the theme toggle is labelled", async () => {
	await render(
		<LocaleProvider locale="en">
			<TabsLayout />
		</LocaleProvider>,
	);
	const header = mockCaptured.screenOptions?.header;
	if (header === undefined) throw new Error("no header");
	await render(header());
	expect(screen.getByRole("header", { name: "HexaScore" })).toBeOnTheScreen();
	await fireEvent.press(screen.getByRole("button", { name: "Open search" }));
	expect(mockNavigate).toHaveBeenCalledWith("/search");
	expect(screen.getByRole("button", { name: /^Switch to (dark|light) theme$/ })).toBeOnTheScreen();
});
