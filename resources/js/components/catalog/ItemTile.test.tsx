import { cleanup, render, screen, within } from "@testing-library/react";
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

// Owner round 2026-11-01 item 22: title and year on the left 2/3, the badge fills the right 1/3.
test("the tile footer holds title and year on the left and a full-height score on the right", () => {
	render(
		<ul>
			<ItemTile item={item} href="/movies/bigbug" />
		</ul>,
	);
	const link = screen.getByRole("link");
	expect(link.getAttribute("href")).toBe("/movies/bigbug");
	const footer = link.querySelector("[data-tile-footer]");
	expect(footer?.className).toContain("grid-cols-[2fr_1fr]");
	expect(footer?.className).toContain("min-h-16");
	const [left, badge] = [...(footer?.children ?? [])];
	expect(within(left as HTMLElement).getByText(item.title).className).toContain("line-clamp-2");
	expect(within(left as HTMLElement).getByText("2022")).toBeTruthy();
	const score = screen.getByRole("img", { name: /score.outOf.*8\.4/ });
	expect(score).toBe(badge);
	expect(score.className).toContain("size-full");
	expect(score.className).toContain("justify-center");
});

test("an item without score or year keeps the footer and the accessible no-score label", () => {
	render(
		<ul>
			<ItemTile item={{ ...item, score: null, year: null }} href="/movies/bigbug" />
		</ul>,
	);
	expect(screen.getByRole("img", { name: "score.none" }).textContent).toBe("–");
	expect(screen.queryByText("2022")).toBeNull();
});
