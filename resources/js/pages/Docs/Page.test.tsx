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

test("every code block gets a copy button that copies the code and confirms", async () => {
	const writeText = vi.fn(() => Promise.resolve());
	Object.defineProperty(navigator, "clipboard", { value: { writeText }, configurable: true });
	render(
		<Page
			slug="pagination"
			title="Pagination"
			body={
				"<h1>Pagination</h1><pre><code>curl https://api.hexascore.app/v1/items</code></pre>"
			}
			toc={[]}
			pages={pages}
		/>,
	);
	const button = screen.getByRole("button", { name: "docs.copy" });
	button.click();
	expect(writeText).toHaveBeenCalledWith("curl https://api.hexascore.app/v1/items");
	expect(await screen.findByRole("button", { name: "docs.copied" })).toBeTruthy();
});
