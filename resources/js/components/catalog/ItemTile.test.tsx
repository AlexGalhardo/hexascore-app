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
	releasePrecision: "day",
	score: 8.4,
	ratingCount: 12,
	cover: null,
};

test("the cover fills a flat card with score and title overlaid", () => {
	render(
		<ul>
			<ItemTile item={item} href="/movies/bigbug" />
		</ul>,
	);
	const link = screen.getByRole("link");
	expect(link.getAttribute("href")).toBe("/movies/bigbug");
	expect(link.className).toContain("rounded-sm");
	expect(link.className).toContain("border-border");
	expect(link.className).not.toContain("shadow-");
	expect(link.querySelector("[data-tile-footer]")).toBeNull();
	expect(screen.getByText(item.title).className).toContain("absolute");
	expect(screen.queryByText("2022")).toBeNull();
	expect(
		screen.getByRole("img", { name: /score.outOf.*8\.4/ }).parentElement?.className,
	).toContain("top-2");
});

test("an item without a score keeps the accessible no-score label", () => {
	render(
		<ul>
			<ItemTile item={{ ...item, score: null, type: "book" }} href="/books/bigbug" />
		</ul>,
	);
	expect(screen.getByRole("img", { name: "score.none" }).textContent).toBe("–");
	expect(screen.getByRole("link").className).toContain("border-border");
});
