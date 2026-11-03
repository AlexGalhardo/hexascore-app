import { afterEach, beforeEach, expect, test, vi } from "vitest";

import { cleanup, fireEvent, render, screen, within } from "../lib/testing";
import AdminLayout from "./AdminLayout";

vi.mock("@inertiajs/react", () => ({
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
