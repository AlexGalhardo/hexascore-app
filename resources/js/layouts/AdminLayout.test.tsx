import { act } from "@testing-library/react";
import { afterEach, beforeEach, expect, test, vi } from "vitest";

import { cleanup, fireEvent, render, screen, within } from "../lib/testing";
import AdminLayout from "./AdminLayout";

const handlers: Record<string, (event?: unknown) => void> = {};

vi.mock("@inertiajs/react", () => ({
	router: {
		on: (name: string, cb: (event?: unknown) => void) => {
			handlers[name] = cb;
			return () => {
				delete handlers[name];
			};
		},
	},
	usePage: () => ({
		url: "/admin/ingest/reviews?x=1",
		props: { appName: "HexaScore", auth: { user: { name: "Ada Lovelace" } } },
	}),
	Link: ({ href, children, ...rest }: { href: string; children: React.ReactNode }) => (
		<a href={href} {...rest}>
			{children}
		</a>
	),
}));
vi.mock("../lib/useT", () => ({
	useT: () => (key: string, p?: Record<string, string>) =>
		p ? `${key} ${JSON.stringify(p)}` : key,
}));

beforeEach(() => {
	document.documentElement.classList.remove("dark");
});
afterEach(cleanup);

test("the sidebar lists the groups and marks only the current screen", () => {
	render(
		<AdminLayout>
			<p>content</p>
		</AdminLayout>,
	);
	const nav = screen.getAllByRole("navigation", { name: "admin.topbar.navLabel" })[0];
	expect(within(nav).getByText("admin.groups.catalogue")).toBeTruthy();
	const current = within(nav).getAllByRole("link", { current: "page" });
	expect(current).toHaveLength(1);
	expect(current[0].getAttribute("href")).toBe("/admin/ingest/reviews");
	expect(screen.getByText("content")).toBeTruthy();
});

test("the theme switch flips the dark class on the document", () => {
	render(
		<AdminLayout>
			<p>content</p>
		</AdminLayout>,
	);
	fireEvent.click(screen.getByRole("button", { name: "admin.topbar.toggleTheme" }));
	expect(document.documentElement.classList.contains("dark")).toBe(true);
	fireEvent.click(screen.getByRole("button", { name: "admin.topbar.toggleTheme" }));
	expect(document.documentElement.classList.contains("dark")).toBe(false);
});

test("the skip link points at the main region", () => {
	render(
		<AdminLayout>
			<p>content</p>
		</AdminLayout>,
	);
	expect(screen.getByRole("link", { name: "admin.topbar.skip" }).getAttribute("href")).toBe(
		"#admin-main",
	);
	expect(document.getElementById("admin-main")).not.toBeNull();
});

const visit = (preserveState: boolean) => ({ detail: { visit: { preserveState } } });

test("a page visit shows the progress bar after a short delay and finish hides it", () => {
	vi.useFakeTimers();
	try {
		render(
			<AdminLayout>
				<p>content</p>
			</AdminLayout>,
		);
		act(() => handlers.start(visit(false)));
		expect(screen.queryByRole("progressbar")).toBeNull();
		act(() => vi.advanceTimersByTime(130));
		expect(screen.getByRole("progressbar", { name: "admin.topbar.loading" })).toBeTruthy();
		expect(document.getElementById("admin-main")?.getAttribute("aria-busy")).toBe("true");
		act(() => handlers.finish());
		expect(screen.queryByRole("progressbar")).toBeNull();
		expect(document.getElementById("admin-main")?.getAttribute("aria-busy")).toBe("false");
	} finally {
		vi.useRealTimers();
	}
});

test("a quick visit never shows the bar, and a reload that keeps the page is ignored", () => {
	vi.useFakeTimers();
	try {
		render(
			<AdminLayout>
				<p>content</p>
			</AdminLayout>,
		);
		act(() => handlers.start(visit(false)));
		act(() => vi.advanceTimersByTime(50));
		act(() => handlers.finish());
		act(() => vi.advanceTimersByTime(300));
		expect(screen.queryByRole("progressbar")).toBeNull();
		act(() => handlers.start(visit(true)));
		act(() => vi.advanceTimersByTime(300));
		expect(screen.queryByRole("progressbar")).toBeNull();
	} finally {
		vi.useRealTimers();
	}
});
