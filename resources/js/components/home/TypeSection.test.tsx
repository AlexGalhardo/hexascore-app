import { cleanup, render, screen, within } from "@testing-library/react";
import type { ReactNode } from "react";
import { afterEach, expect, test, vi } from "vitest";

import type { Highlight, Section } from "../catalog/SectionHighlights";
import TypeSection, { pickRandom, type TrendingEntry } from "./TypeSection";

afterEach(cleanup);
// jsdom has no matchMedia; reduced motion keeps the carousel still in these tests.
vi.stubGlobal("matchMedia", () => ({
	matches: true,
	addEventListener: () => {},
	removeEventListener: () => {},
}));

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

// 20 pool items, two of them also trending.
const pool = [h("t1"), h("t2"), ...Array.from({ length: 18 }, (_, i) => h(`p${i}`))];
const section: Section = { type: "game", count: 24873, pool };

const trending: TrendingEntry[] = [
	{ ...h("t1", 8.4), rank: 1, basis: "trend", genre: "Action" },
	{ ...h("t2"), rank: 2, basis: "trend", genre: null },
	{ ...h("t3"), rank: 3, basis: "prior", genre: "Adventure" },
];

// Owner round 2026-11-01 item 26 (b): a random 12 of the pool, never a trending item, picked after hydration.
test("pickRandom returns n unique items driven by the injected rng", () => {
	const items = Array.from({ length: 36 }, (_, i) => i);
	const seq = [0.99, 0, 0.5, 0.25];
	let k = 0;
	const rng = () => seq[k++ % seq.length];
	const got = pickRandom(items, 12, rng);
	expect(got).toHaveLength(12);
	expect(new Set(got).size).toBe(12);
	expect(got[0]).toBe(35);
	expect(got[1]).toBe(1);
	expect(pickRandom(items, 12, () => 0)).toEqual(items.slice(0, 12));
	expect(pickRandom([1, 2], 12)).toHaveLength(2);
});

test("the More grid shows 12 tiles from the pool without the trending ones", () => {
	render(<TypeSection section={section} trending={trending} />);
	const grid = screen.getByText("home.more.game").nextElementSibling as HTMLElement;
	const hrefs = [...grid.querySelectorAll("a")].map((a) => a.getAttribute("href"));
	expect(hrefs).toHaveLength(12);
	expect(new Set(hrefs).size).toBe(12);
	expect(hrefs).not.toContain("/games/slug-t1");
	expect(hrefs).not.toContain("/games/slug-t2");
	expect(grid.className).toContain("md:grid-cols-6");
});

test("trending cards with rank, score explanation, then more titles and see all", () => {
	render(<TypeSection section={section} trending={trending} />);
	const carousel = screen.getByRole("region", { name: /home.carousel.label/ });
	const cards = within(carousel).getAllByRole("listitem").slice(0, 3);
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
	const empty = { ...section, pool: [] };
	const { container } = render(<TypeSection section={empty} trending={[]} />);
	expect(container.innerHTML).toBe("");
});
