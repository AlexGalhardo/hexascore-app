import userEvent from "@testing-library/user-event";
import type { ReactNode } from "react";
import { afterEach, expect, test, vi } from "vitest";
import { cleanup, render, screen, within } from "../lib/testing";

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
	const footer = screen.getByRole("navigation", { name: "footer.label" });
	expect(footer.querySelectorAll("a")).toHaveLength(6);
	expect(screen.getByRole("button", { name: "footer.cookieSettings" })).toBeTruthy();
	expect(footer.querySelector("[hreflang]")).toBeNull();
});

test("each section link starts with its category icon and keeps the name as text", () => {
	render(<Nav />);
	const sections = screen.getByRole("navigation", { name: "nav.sections" });
	const links = within(sections).getAllByRole("link");
	expect(links.map((l) => l.firstElementChild?.getAttribute("data-category-icon"))).toEqual([
		"game",
		"movie",
		"series",
		"book",
	]);
	expect(links[0].textContent).toBe("nav.games");
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

// jsdom has no layout: the classes are the contract. At lg (1024 px) four sections
// plus Feed pushed the bell off-screen, so Feed sits inline only from xl.
test("the inline Feed link waits for xl and the account menu carries it below", () => {
	try {
		props.auth.user = { name: "Ana", nickname: "ana" };
		props.flags = { social: true };
		render(<Nav />);
		const [sections] = screen.getAllByRole("navigation", { name: "nav.sections" });
		const inline = sections.querySelector('a[href="/feed"]');
		expect(inline?.className).toMatch(/(^|\s)hidden(\s|$)/);
		// a bare inline-flex would win over hidden in Tailwind's sort and undo the fix
		expect(inline?.className).not.toMatch(/(^|\s)inline-flex(\s|$)/);
		expect(inline?.className).toContain("xl:inline-flex");
		// The trigger is the avatar alone (owner round 2026-11-01 item 13).
		const trigger = screen.getAllByRole("button", { name: "nav.account" })[0];
		expect(trigger.querySelector(".xl\\:inline")).toBeNull();
	} finally {
		props.auth.user = null;
		props.flags = {};
	}
});

// Owner round 2026-11-01 item 13: no user name in the account dropdown, an icon before each link.
test("the account dropdown shows no user name and every entry starts with an icon", async () => {
	try {
		props.auth.user = { name: "Ana Real", nickname: "ana", role: "admin" };
		props.flags = { social: true };
		render(<Nav />);
		const trigger = screen.getAllByRole("button", { name: "nav.account" })[0];
		expect(trigger.textContent).toBe("A");
		await userEvent.click(trigger);
		const panel = document.getElementById(trigger.getAttribute("aria-controls") ?? "");
		expect(panel).toBeTruthy();
		expect(panel?.textContent).not.toContain("Ana Real");
		const entries = [...(panel?.querySelectorAll("a, button") ?? [])];
		expect(entries.length).toBeGreaterThanOrEqual(7);
		for (const entry of entries) expect(entry.firstElementChild?.tagName).toBe("svg");
	} finally {
		props.auth.user = null;
		props.flags = {};
	}
});

// Owner, 2026-11-01: below lg the bar holds only the title, the search field and a hamburger;
// everything else lives in the hamburger's sheet (supersedes round 2 item 7 on web mobile).
test("below lg the bar keeps title, search field and hamburger; the rest is desktop-only", () => {
	render(<Nav />);
	const search = screen.getByRole("link", { name: "nav.openSearch" });
	expect(search.className).toContain("flex-1");
	expect(search.querySelector("span")?.className).not.toMatch(/(^|\s)hidden(\s|$)/);
	const desktop = screen.getByRole("button", { name: "nav.language" }).closest("[data-desktop]");
	expect(desktop?.className).toMatch(/(^|\s)hidden(\s|$)/);
	expect(desktop?.className).toContain("lg:flex");
	expect(screen.getAllByRole("navigation", { name: "nav.sections" })).toHaveLength(1);
	const burger = screen.getByRole("button", { name: "nav.menu" });
	expect(burger.className).toContain("lg:hidden");
	expect(burger.className).toContain("size-11");
	expect(burger.getAttribute("aria-expanded")).toBe("false");
});

// Kept before the other sheet tests so it covers the very first open, while the lazy chunk still loads.
test("the lazy sheet is not in the first render and the first open still traps Tab", async () => {
	const user = userEvent.setup({ delay: null });
	render(<Nav />);
	expect(screen.queryByRole("dialog")).toBeNull();
	const burger = screen.getByRole("button", { name: "nav.menu" });
	await user.click(burger);
	// Keys pressed before the sheet arrives must not reach the page behind it.
	for (let i = 0; i < 5; i++) {
		await user.keyboard("{Tab}");
		const sheet = screen.queryByRole("dialog");
		const a = document.activeElement;
		expect(a === burger || !!sheet?.contains(a)).toBe(true);
	}
	const sheet = await screen.findByRole("dialog", {}, { timeout: 5000 });
	for (const keys of ["{Tab}", "{Shift>}{Tab}{/Shift}"]) {
		for (let i = 0; i < 20; i++) {
			await user.keyboard(keys);
			expect(sheet.contains(document.activeElement)).toBe(true);
		}
	}
	await user.keyboard("{Escape}");
	expect(screen.queryByRole("dialog")).toBeNull();
	expect(document.activeElement).toBe(burger);
});

test("the hamburger sheet holds sections, language, theme and login; Escape returns focus", async () => {
	const user = userEvent.setup();
	render(<Nav />);
	const burger = screen.getByRole("button", { name: "nav.menu" });
	await user.click(burger);
	expect(burger.getAttribute("aria-expanded")).toBe("true");
	const sheet = await screen.findByRole("dialog", {}, { timeout: 5000 });
	expect(sheet.contains(document.activeElement)).toBe(true);
	const hrefs = [...sheet.querySelectorAll("a")].map((a) => a.getAttribute("href"));
	expect(hrefs).toEqual(
		expect.arrayContaining(["/games", "/movies", "/series", "/books", "/login"]),
	);
	expect(sheet.querySelectorAll("a[hreflang]")).toHaveLength(3);
	expect(sheet.querySelector('[aria-current="true"]')?.getAttribute("hreflang")).toBe("en");
	expect(sheet.querySelector("button[data-theme-toggle]")).not.toBeNull();
	expect(sheet.className).toContain("motion-reduce:transition-none");

	await user.keyboard("{Escape}");
	expect(screen.queryByRole("dialog")).toBeNull();
	expect(burger.getAttribute("aria-expanded")).toBe("false");
	expect(document.activeElement).toBe(burger);
});

test("signed in with the flag on, the hamburger shows unread and the sheet has feed and account", async () => {
	const user = userEvent.setup();
	try {
		props.auth.user = { name: "Ana", nickname: "ana" };
		props.flags = { social: true };
		props.unreadNotifications = 3;
		render(<Nav />);
		const burger = screen.getByRole("button", { name: "nav.menuUnread" });
		expect(burger.textContent).toContain("3");
		await user.click(burger);
		const sheet = await screen.findByRole("dialog", {}, { timeout: 5000 });
		const hrefs = [...sheet.querySelectorAll("a")].map((a) => a.getAttribute("href"));
		expect(hrefs).toEqual(
			expect.arrayContaining([
				"/feed",
				"/me/notifications",
				"/library",
				"/lists",
				"/u/ana",
				"/logout",
			]),
		);
		expect(sheet.textContent).toContain("Ana");
		expect(sheet.querySelector('a[href="/me/notifications"]')?.textContent).toContain("3");
	} finally {
		props.auth.user = null;
		props.flags = {};
		props.unreadNotifications = 0;
	}
});

// Base UI's focus guards refocus in a requestAnimationFrame, so a Tab that lands before the next
// frame (a loaded main thread, or keys pressed right after opening) used to escape to the page.
test("Tab and Shift+Tab right after opening never leave the sheet", async () => {
	for (const keys of ["{Tab}", "{Shift>}{Tab}{/Shift}"]) {
		const user = userEvent.setup({ delay: null });
		render(<Nav />);
		await user.click(screen.getByRole("button", { name: "nav.menu" }));
		const sheet = await screen.findByRole("dialog", {}, { timeout: 5000 });
		const outside: string[] = [];
		for (let i = 0; i < 40; i++) {
			await user.keyboard(keys);
			const a = document.activeElement;
			if (!sheet.contains(a))
				outside.push(`${i}:${a?.tagName}:${a?.textContent?.slice(0, 20)}`);
		}
		expect(outside).toEqual([]);
		cleanup();
	}
});

test("the hamburger name follows the unread count when it changes", () => {
	try {
		props.auth.user = { name: "Ana", nickname: "ana" };
		const { rerender } = render(<Nav />);
		expect(screen.getByRole("button", { name: "nav.menu" })).toBeTruthy();
		props.unreadNotifications = 2;
		rerender(<Nav />);
		const burger = screen.getByRole("button", { name: "nav.menuUnread" });
		expect(burger.textContent).toContain("2");
	} finally {
		props.auth.user = null;
		props.unreadNotifications = 0;
	}
});

test("the search palette is loaded on the first Ctrl+K, not with the navbar", async () => {
	render(<Nav />);
	expect(screen.queryByRole("dialog")).toBeNull();
	await userEvent.keyboard("{Control>}k{/Control}");
	// The palette chunk is imported on demand; under a loaded host that takes longer than the 1 s default.
	expect(await screen.findByRole("dialog", {}, { timeout: 5000 })).toBeTruthy();
});
