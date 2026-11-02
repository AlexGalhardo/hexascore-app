import { cleanup, render, screen } from "@testing-library/react";
import type { ReactNode } from "react";
import { afterEach, expect, test, vi } from "vitest";

import ItemTile from "./ItemTile";
import type { Highlight } from "./SectionHighlights";

afterEach(cleanup);

vi.mock("../../lib/useT", () => ({
	useT: () => (key: string, p?: Record<string, string | number>) =>
		p ? `${key} ${JSON.stringify(p)}` : key,
}));
vi.mock("@inertiajs/react", () => ({
	Link: ({
		href,
		children,
		className,
	}: {
		href: string;
		children: ReactNode;
		className?: string;
	}) => (
		<a href={href} className={className}>
			{children}
		</a>
	),
	usePage: () => ({ props: { locale: "en", coversBase: "/c/" } }),
}));

const item: Highlight = {
	id: "1",
	type: "movie",
	slug: "bigbug",
	title: "A title long enough to need a second line on a narrow tile",
	year: 2022,
	score: 8.4,
	ratingCount: 12,
	cover: null,
};

// Owner decisions 2026-11-02: no year, title and score on one row, category-coloured border.
test("the tile footer holds the truncating title and the score on one row, with a category border", () => {
	render(
		<ul>
			<ItemTile item={item} href="/movies/bigbug" />
		</ul>,
	);
	const link = screen.getByRole("link");
	expect(link.getAttribute("href")).toBe("/movies/bigbug");
	expect(link.className).toContain("border-cat-movie");
	const footer = link.querySelector("[data-tile-footer]");
	expect(footer?.className).toContain("grid-cols-[1fr_auto]");
	const [title, badge] = [...(footer?.children ?? [])];
	expect(title.textContent).toBe(item.title);
	expect(title.className).toContain("truncate");
	expect(screen.queryByText("2022")).toBeNull();
	expect(screen.getByRole("img", { name: /score.outOf.*8\.4/ })).toBe(badge);
});

test("an item without a score keeps the footer and the accessible no-score label", () => {
	render(
		<ul>
			<ItemTile item={{ ...item, score: null, type: "book" }} href="/books/bigbug" />
		</ul>,
	);
	expect(screen.getByRole("img", { name: "score.none" }).textContent).toBe("–");
	expect(screen.getByRole("link").className).toContain("border-cat-book");
});
