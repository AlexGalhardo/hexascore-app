import { cleanup, render, screen, within } from "@testing-library/react";
import type { ReactNode } from "react";
import { afterEach, expect, test, vi } from "vitest";

import type { Highlight, Section } from "../catalog/SectionHighlights";
import TypeSection from "./TypeSection";

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

const h = (id: string): Highlight => ({
	id,
	type: "game",
	slug: `slug-${id}`,
	title: `Title ${id}`,
	year: 2027,
	releaseDate: "2027-09-09",
	releasePrecision: "day",
	score: null,
	ratingCount: 0,
	cover: null,
});
const section: Section = {
	type: "game",
	count: 24873,
	pool: Array.from({ length: 14 }, (_, i) => h(`released-${i}`)),
};
const upcoming = Array.from({ length: 12 }, (_, i) => h(`upcoming-${i}`));
upcoming[1] = { ...upcoming[1], releaseDate: "2027-01-01", releasePrecision: "month" };

test("upcoming carousel and released trending grid use separate ordered lists", () => {
	render(<TypeSection section={section} upcoming={upcoming} />);
	const carousel = screen
		.getAllByRole("region", { name: /home.mostWaited/ })
		.find(
			(region) => region.getAttribute("aria-roledescription") === "carousel",
		) as HTMLElement;
	expect(within(carousel).getAllByRole("listitem")).toHaveLength(6);
	expect(
		within(carousel)
			.getByRole("link", { name: /Title upcoming-0/ })
			.getAttribute("href"),
	).toBe("/games/slug-upcoming-0");
	expect(
		within(within(carousel).getByRole("link", { name: /Title upcoming-0/ })).getByText(
			"09 September 2027",
		),
	).toBeTruthy();
	expect(within(carousel).getByText("January 2027")).toBeTruthy();
	expect(within(carousel).queryByRole("img", { name: /score\./ })).toBeNull();
	const grid = screen.getByText(/home.trending30Days/).nextElementSibling as HTMLElement;
	const links = [...grid.querySelectorAll("a")];
	expect(links).toHaveLength(12);
	expect(links[0].getAttribute("href")).toBe("/games/slug-released-0");
	expect(links[11].getAttribute("href")).toBe("/games/slug-released-11");
});

test("the count appears in See all and not as a separate heading number", () => {
	render(<TypeSection section={section} upcoming={upcoming} />);
	const header = document.querySelector("[data-section-header]") as HTMLElement;
	expect(header.children).toHaveLength(2);
	expect(within(header).getByRole("heading", { level: 2 }).textContent).toMatch(
		/home.mostWaited/,
	);
	expect(within(header).getByRole("link", { name: /home.seeAll.game/ }).textContent).toMatch(
		/24,873/,
	);
});

test("a type with neither upcoming nor released trending items renders nothing", () => {
	const { container } = render(<TypeSection section={{ ...section, pool: [] }} />);
	expect(container.innerHTML).toBe("");
});
