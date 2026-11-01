import { cleanup, render, screen } from "@testing-library/react";
import type { ReactNode } from "react";
import { afterEach, expect, test, vi } from "vitest";

import Page from "./Page";

vi.mock("../../lib/useT", () => ({ useT: () => (key: string) => key }));
vi.mock("../../components/Seo", () => ({ default: () => null }));
vi.mock("@inertiajs/react", () => ({
	Link: ({ children, ...rest }: { href: string; children: ReactNode }) => (
		<a {...rest}>{children}</a>
	),
	usePage: () => ({ props: { locale: "en" } }),
}));

afterEach(cleanup);

const pages = ["index", "authentication", "pagination"];

test("marks the current page and links the previous and next pages", () => {
	render(
		<Page
			slug="authentication"
			title="Authentication"
			body="<h1>Authentication</h1>"
			toc={[{ id: "bearer", text: "Bearer token", level: 2 }]}
			pages={pages}
		/>,
	);
	const current = screen.getAllByRole("link", { name: "docs.nav.authentication" })[0];
	expect(current.getAttribute("aria-current")).toBe("page");
	const links = screen.getAllByRole("link").map((a) => a.getAttribute("href"));
	expect(links).toContain("#bearer");
	expect(
		screen
			.getAllByRole("link", { name: /docs\.nav\.index/ })
			.at(-1)
			?.getAttribute("href"),
	).toBe("/docs");
	expect(
		screen
			.getAllByRole("link", { name: /docs\.nav\.pagination/ })
			.at(-1)
			?.getAttribute("href"),
	).toBe("/docs/pagination");
});

test("the first page has no previous link", () => {
	render(<Page slug="index" title="Quickstart" body="" toc={[]} pages={pages} />);
	expect(screen.getAllByRole("link", { name: /docs\.nav\.index/ })).toHaveLength(1);
});
