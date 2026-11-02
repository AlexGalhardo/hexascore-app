import { cleanup, render, screen } from "@testing-library/react";
import type { ReactNode } from "react";
import { afterEach, expect, test, vi } from "vitest";

import { pageCanonical } from "../../pages/Catalog/Index";
import Pager, { pageHref } from "./Pager";

vi.mock("../../lib/useT", () => ({ useT: () => (key: string) => key }));
vi.mock("@inertiajs/react", () => ({
	Head: ({ children }: { children: ReactNode }) => <>{children}</>,
	Link: ({ href, rel, children }: { href: string; rel?: string; children: ReactNode }) => (
		<a href={href} rel={rel}>
			{children}
		</a>
	),
	usePage: () => ({ url: "/", props: {} }),
}));

afterEach(cleanup);

test("pageHref sets a readable page number, drops page 1 and old cursors, keeps the filters", () => {
	expect(pageHref("/pt-br/games?genre=rpg&page=2", "3")).toBe("/pt-br/games?genre=rpg&page=3");
	expect(pageHref("/games?page=2&genre=rpg", "1")).toBe("/games?genre=rpg");
	expect(pageHref("/games?cursor=abc&genre=rpg", "2")).toBe("/games?genre=rpg&page=2");
});

test("prev and next are links with rel; a missing end is disabled", () => {
	render(<Pager url="/games?genre=rpg" pagination={{ prev: "", next: "2" }} label="Pages" />);
	expect(screen.getByRole("navigation", { name: "Pages" })).toBeTruthy();
	const next = screen.getByRole("link", { name: "catalog.next" });
	expect(next.getAttribute("href")).toBe("/games?genre=rpg&page=2");
	expect(next.getAttribute("rel")).toBe("next");
	expect(screen.getByText("catalog.previous").getAttribute("aria-disabled")).toBe("true");
});

test("one page renders no pager", () => {
	const { container } = render(<Pager url="/games" pagination={{ prev: "", next: "" }} />);
	expect(container.innerHTML).toBe("");
});

test("the canonical keeps the page number and a non-default page size, never the filters", () => {
	expect(pageCanonical("https://hexascore.app/games", "/games?genre=rpg&page=3", 40)).toBe(
		"https://hexascore.app/games?page=3",
	);
	expect(pageCanonical("https://hexascore.app/games", "/games?page=1&cursor=x", 40)).toBe(
		"https://hexascore.app/games",
	);
	expect(pageCanonical("https://hexascore.app/games", "/games?genre=rpg&per=80", 80)).toBe(
		"https://hexascore.app/games?per=80",
	);
	expect(pageCanonical("https://hexascore.app/games", "/games?per=7", 40)).toBe(
		"https://hexascore.app/games",
	);
});

test("page links keep the page size", () => {
	expect(pageHref("/games?per=120&page=2", "3")).toBe("/games?per=120&page=3");
});

test("removing a filter chip drops that parameter and the page number, keeping the others", async () => {
	const { withoutParam } = await import("../../pages/Catalog/Index");
	expect(withoutParam("/pt-br/games?company=big-studio&genre=rpg&page=3", "company")).toBe(
		"/pt-br/games?genre=rpg",
	);
});
