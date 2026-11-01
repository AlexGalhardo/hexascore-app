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

test("pageHref swaps the cursor and keeps the filters", () => {
	expect(pageHref("/pt-br/games?genre=rpg&cursor=a", "b")).toBe(
		"/pt-br/games?genre=rpg&cursor=b",
	);
	expect(pageHref("/games?cursor=a&genre=rpg", "")).toBe("/games?genre=rpg");
});

test("prev and next are links with rel; a missing end is disabled", () => {
	render(<Pager url="/games?genre=rpg" pagination={{ prev: "", next: "n1" }} label="Pages" />);
	expect(screen.getByRole("navigation", { name: "Pages" })).toBeTruthy();
	const next = screen.getByRole("link", { name: "catalog.next" });
	expect(next.getAttribute("href")).toBe("/games?genre=rpg&cursor=n1");
	expect(next.getAttribute("rel")).toBe("next");
	expect(screen.getByText("catalog.previous").getAttribute("aria-disabled")).toBe("true");
});

test("one page renders no pager", () => {
	const { container } = render(<Pager url="/games" pagination={{ prev: "", next: "" }} />);
	expect(container.innerHTML).toBe("");
});

test("the canonical keeps the cursor and drops the filters", () => {
	expect(pageCanonical("https://hexascore.app/games", "/games?genre=rpg&cursor=c1")).toBe(
		"https://hexascore.app/games?cursor=c1",
	);
	expect(pageCanonical("https://hexascore.app/games", "/games?genre=rpg")).toBe(
		"https://hexascore.app/games",
	);
});
