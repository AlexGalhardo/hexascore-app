import { cleanup, render, screen, within } from "@testing-library/react";
import type { ReactNode } from "react";
import { afterEach, expect, test, vi } from "vitest";

import type { Highlight, Section } from "../catalog/SectionHighlights";
import TypeSection, { otherTitles, type TrendingEntry } from "./TypeSection";

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

const h = (id: string, score: number | null = null): Highlight => ({
	id,
	type: "game",
	slug: `slug-${id}`,
	title: `Title ${id}`,
	year: 2020,
	score,
	ratingCount: 3,
	cover: null,
});

const section: Section = {
	type: "game",
	count: 24873,
	topRated: [h("a"), h("b"), h("t1")],
	newest: [h("c"), h("d"), h("e"), h("f"), h("g")],
	mostRated: [h("t2"), h("b"), h("h")],
};

const trending: TrendingEntry[] = [
	{ ...h("t1", 8.4), rank: 1, basis: "trend", genre: "Action" },
	{ ...h("t2"), rank: 2, basis: "trend", genre: null },
	{ ...h("t3"), rank: 3, basis: "prior", genre: "Adventure" },
];

test("other titles skip trending ones and duplicates, six at most", () => {
	expect(otherTitles(section, trending).map((i) => i.id)).toEqual(["b", "h", "a", "c", "d", "e"]);
});

test("three trending cards with rank, score explanation, then more titles and see all", () => {
	render(<TypeSection section={section} trending={trending} />);
	const cards = screen.getAllByRole("listitem").slice(0, 3);
	expect(within(cards[0]).getByText(/home.trendingRank.*"rank":1/)).toBeTruthy();
	expect(within(cards[0]).getByText("item.band.high")).toBeTruthy();
	expect(within(cards[1]).getByText(/home.ratingsOfMin.*"count":3,"min":5/)).toBeTruthy();
	// Owner round 2026-11-01 item 15: "#3 trending" even on the prior-week fallback, and the genre, not the type.
	expect(within(cards[2]).getByText(/home.trendingRank.*"rank":3/)).toBeTruthy();
	expect(within(cards[0]).getByText("2020 · Action")).toBeTruthy();
	expect(within(cards[1]).getByText("2020")).toBeTruthy();
	expect(screen.queryByText(/itemTypes\./)).toBeNull();
	expect(cards[0].querySelector("a")?.getAttribute("href")).toBe("/games/slug-t1?from=trending");
	expect(screen.getByText("home.more.game")).toBeTruthy();
	expect(
		screen
			.getByText(/home.seeAll.game/)
			.closest("a")
			?.getAttribute("href"),
	).toBe("/games");
	expect(screen.getByText(/24,873/)).toBeTruthy();
	const heading = screen.getByRole("heading", { level: 2 });
	expect(
		heading.parentElement?.querySelector('[data-category-icon="game"][aria-hidden="true"]'),
	).toBeTruthy();
});

// Owner round 2026-11-01 item 23: name left, larger count centred, "see all" right; stacked below sm.
test("the section header is a 1fr auto 1fr grid with the locale-formatted count in the middle", () => {
	render(<TypeSection section={section} trending={trending} />);
	const header = document.querySelector("[data-section-header]") as HTMLElement;
	expect(header.className).toContain("sm:grid-cols-[1fr_auto_1fr]");
	expect(header.className).not.toMatch(/(^|\s)grid-cols-/);
	const [left, middle, right] = [...header.children] as HTMLElement[];
	expect(within(left).getByRole("heading", { level: 2 }).textContent).toBe("catalog.title.game");
	expect(middle.textContent).toMatch(/home\.titles.*"n":"24,873"/);
	expect(middle.className).toContain("text-xl");
	expect(middle.className).toContain("sm:text-center");
	expect(right.tagName).toBe("A");
	expect(right.className).toContain("sm:justify-self-end");
});

test("a type with nothing to show renders nothing", () => {
	const empty = { ...section, topRated: [], newest: [], mostRated: [] };
	const { container } = render(<TypeSection section={empty} trending={[]} />);
	expect(container.innerHTML).toBe("");
});
