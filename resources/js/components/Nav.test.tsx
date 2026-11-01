import { cleanup, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { ReactNode } from "react";
import { afterEach, expect, test, vi } from "vitest";

import Footer from "./Footer";
import Nav from "./Nav";
import Btn, { btnClass } from "./ui/Btn";

const props = vi.hoisted(() => ({
	appName: "HexaScore",
	locale: "en",
	auth: { user: null } as { user: null | Record<string, unknown> },
	unreadNotifications: 0,
	flags: {} as { social?: boolean },
}));

vi.mock("@inertiajs/react", () => ({
	usePage: () => ({ url: "/games", props }),
	Link: ({ children, href, ...rest }: { children: ReactNode; href: string }) => (
		<a href={href} {...rest}>
			{children}
		</a>
	),
}));
vi.mock("../lib/useT", () => ({ useT: () => (key: string) => key }));

afterEach(cleanup);

test("btnClass gives every kind a 44 px target, a ledge and a reduced-motion press", () => {
	for (const kind of ["primary", "secondary"] as const) {
		const cls = btnClass(kind);
		expect(cls).toContain("min-h-11");
		expect(cls).toContain(`shadow-ledge-${kind}`);
		expect(cls).toContain(`active:shadow-ledge-${kind}-pressed`);
		expect(cls).toContain("active:translate-y-px");
		expect(cls).toContain("motion-reduce:active:translate-y-0");
		expect(cls).toContain("focus-visible:outline-2");
	}
	expect(btnClass("ghost")).not.toContain("shadow-ledge");
	render(<Btn kind="secondary">Reset</Btn>);
	expect(screen.getByRole("button", { name: "Reset" }).getAttribute("type")).toBe("button");
});

test("language dropdown opens, marks the current locale and closes on Escape", async () => {
	const user = userEvent.setup();
	render(<Nav />);
	const trigger = screen.getByRole("button", { name: "nav.language" });
	expect(trigger.getAttribute("aria-expanded")).toBe("false");
	expect(screen.queryByRole("list", { name: "nav.language" })).toBeNull();

	await user.click(trigger);
	const list = screen.getByRole("list", { name: "nav.language" });
	const links = list.querySelectorAll("a");
	expect(links).toHaveLength(3);
	expect(list.querySelector('[aria-current="true"]')?.getAttribute("hreflang")).toBe("en");
	expect(list.querySelector('[hreflang="es"]')?.getAttribute("href")).toBe("/es/games");

	await user.keyboard("{Escape}");
	expect(trigger.getAttribute("aria-expanded")).toBe("false");
	expect(document.activeElement).toBe(trigger);
});

test("navbar has the search trigger, moon/sun toggle and four sections; footer has no languages", () => {
	render(
		<>
			<Nav />
			<Footer />
		</>,
	);
	const search = screen.getByRole("link", { name: "nav.openSearch" });
	expect(search.getAttribute("href")).toBe("/search");
	expect(search.hasAttribute("data-search-trigger")).toBe(true);
	expect(screen.getByRole("button", { name: "nav.theme" }).querySelectorAll("svg")).toHaveLength(
		2,
	);
	expect(screen.getAllByRole("navigation", { name: "nav.sections" })[1].className).toContain(
		"justify-between",
	);
	const footer = screen.getByRole("navigation", { name: "footer.label" });
	expect(footer.querySelectorAll("a")).toHaveLength(6);
	expect(footer.querySelector("[hreflang]")).toBeNull();
});

test("the Feed link shows only to signed-in users while the social flag is on", () => {
	const feedLinks = () => screen.queryAllByRole("link", { name: "social.feed.title" });
	try {
		props.flags = { social: true };
		render(<Nav />);
		expect(feedLinks()).toHaveLength(0);
		cleanup();

		props.auth.user = { name: "Ana", nickname: "ana" };
		props.flags = {};
		render(<Nav />);
		expect(feedLinks()).toHaveLength(0);
		cleanup();

		props.flags = { social: true };
		render(<Nav />);
		expect(feedLinks().map((a) => a.getAttribute("href"))).toEqual(["/feed"]);
	} finally {
		props.auth.user = null;
		props.flags = {};
	}
});
