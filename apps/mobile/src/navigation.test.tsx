import { render, screen } from "@testing-library/react-native";
import type { ReactNode } from "react";
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

test("each SDD 10 route renders its placeholder", async () => {
	await render(<Login />);
	expect(screen.getByText("Login")).toBeTruthy();

	await render(<Home />);
	expect(screen.getByText("HexaScore")).toBeTruthy();

	params({ type: "games" });
	await render(<SectionScreen />);
	expect(screen.getByText("Section games")).toBeTruthy();

	await render(<Search />);
	expect(screen.getByText("Search")).toBeTruthy();

	await render(<Lists />);
	expect(screen.getByText("Lists")).toBeTruthy();

	await render(<Profile />);
	expect(screen.getByText("Profile")).toBeTruthy();

	params({ type: "movies", slug: "arrival" });
	await render(<ItemScreen />);
	expect(screen.getByText("Item movies arrival")).toBeTruthy();

	params({ id: "42" });
	await render(<EntityScreen />);
	expect(screen.getByText("Entity 42")).toBeTruthy();

	params({ id: "7" });
	await render(<ListScreen />);
	expect(screen.getByText("List 7")).toBeTruthy();

	await render(<UpdateRequired />);
	expect(screen.getByText(`Update required ${storeListingUrl()}`)).toBeTruthy();

	await render(<RootLayout />);
	await render(<TabsLayout />);
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
