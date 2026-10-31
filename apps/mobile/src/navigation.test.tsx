import { render, screen } from "@testing-library/react-native";
import type { ReactElement, ReactNode } from "react";
import RootLayout from "../app/_layout";
import Login from "../app/(auth)/login";
import TabsLayout from "../app/(tabs)/_layout";
import SectionScreen, { SectionBody } from "../app/(tabs)/[type]";
import Home from "../app/(tabs)/index";
import Lists from "../app/(tabs)/lists";
import Profile from "../app/(tabs)/profile";
import Search from "../app/(tabs)/search";
import ItemScreen, { ItemBody } from "../app/[type]/[slug]";
import EntityScreen from "../app/entities/[id]";
import ListScreen from "../app/lists/[id]";
import UpdateRequired from "../app/update-required";
import { getNetworkState, setNetworkState } from "./platform/network";
import { storeListingUrl } from "./platform/store";

jest.mock("expo-router", () => {
	const React = require("react") as typeof import("react");
	const { View: MockView } = require("react-native") as typeof import("react-native");
	function Frame({ children }: { children?: ReactNode }): ReactNode {
		return React.createElement(MockView, null, children);
	}
	function Tabs(props: { children?: ReactNode }): ReactNode {
		return Frame(props);
	}
	Tabs.Screen = function Screen(): null {
		return null;
	};
	function Stack(props: { children?: ReactNode }): ReactNode {
		return Frame(props);
	}
	Stack.Screen = function Screen(): null {
		return null;
	};
	return {
		Stack,
		Tabs,
		useLocalSearchParams: () =>
			(globalThis as { __routeParams?: Record<string, string> }).__routeParams ?? {},
	};
});

function params(next: Record<string, string>): void {
	(globalThis as { __routeParams?: Record<string, string> }).__routeParams = next;
}

// One test per route: a single test rendering all eleven screens shared one 5 s
// budget, and the cold first render (lazy nativewind runtime) ate most of it
// when Jest workers competed for CPU.
const routes: [string, Record<string, string>, () => ReactElement, string | null][] = [
	["(auth)/login", {}, () => <Login />, "Login"],
	["(tabs)/index", {}, () => <Home />, "HexaScore"],
	["(tabs)/[type]", { type: "games" }, () => <SectionScreen />, "Section games"],
	["(tabs)/search", {}, () => <Search />, "Search"],
	["(tabs)/lists", {}, () => <Lists />, "Lists"],
	["(tabs)/profile", {}, () => <Profile />, "Profile"],
	[
		"[type]/[slug]",
		{ type: "movies", slug: "arrival" },
		() => <ItemScreen />,
		"Item movies arrival",
	],
	["entities/[id]", { id: "42" }, () => <EntityScreen />, "Entity 42"],
	["lists/[id]", { id: "7" }, () => <ListScreen />, "List 7"],
	["update-required", {}, () => <UpdateRequired />, `Update required ${storeListingUrl()}`],
	["_layout", {}, () => <RootLayout />, null],
	["(tabs)/_layout", {}, () => <TabsLayout />, null],
];

test.each(routes)("route %s renders its placeholder", async (_name, routeParams, element, text) => {
	params(routeParams);
	await render(element());
	if (text !== null) expect(screen.getByText(text)).toBeTruthy();
});

test.each(["games", "movies", "series", "books"])("section and item accept %s", async (type) => {
	await render(<SectionBody type={type} />);
	expect(screen.getByText(`Section ${type}`)).toBeTruthy();
	await render(<ItemBody type={type} slug="title" />);
	expect(screen.getByText(`Item ${type} title`)).toBeTruthy();
});

test("[type] rejects anything outside games|movies|series|books", async () => {
	await render(<SectionBody type="people" />);
	expect(screen.getByText("Unknown section")).toBeTruthy();
	await render(<ItemBody type="billing" slug="nope" />);
	expect(screen.getByText("Unknown section")).toBeTruthy();
	params({ type: "admin" });
	await render(<SectionScreen />);
	expect(screen.getByText("Unknown section")).toBeTruthy();
});

test("platform boundary exposes network state and the Play listing", () => {
	expect(getNetworkState()).toBe("unknown");
	setNetworkState("offline");
	expect(getNetworkState()).toBe("offline");
	setNetworkState("unknown");
	expect(storeListingUrl()).toBe("market://details?id=app.hexascore");
});
