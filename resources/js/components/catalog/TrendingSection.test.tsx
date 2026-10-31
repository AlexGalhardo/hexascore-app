import { cleanup, render, screen } from "@testing-library/react";
import { afterEach, expect, test, vi } from "vitest";

import TrendingSection, { type TrendingEntry } from "./TrendingSection";

afterEach(cleanup);
vi.mock("@inertiajs/react", () => ({
	Link: ({ href, children, ...rest }: { href: string; children: React.ReactNode }) => (
		<a href={href} {...rest}>
			{children}
		</a>
	),
	usePage: () => ({ props: { locale: "en" } }),
}));
vi.mock("../../lib/route", () => ({ localized: (path: string) => path }));
vi.mock("../../lib/useT", () => ({
	useT: () => (key: string, params?: Record<string, string>) =>
		params?.type ? `${key} ${params.type}` : key,
}));

const entry = (rank: number, basis: TrendingEntry["basis"]): TrendingEntry => ({
	id: `id-${rank}`,
	type: "game",
	slug: `game-${rank}`,
	title: `Game ${rank}`,
	year: 2020,
	score: 8.7,
	ratingCount: 9,
	cover: null,
	rank,
	basis,
});

test("renders nothing for a type without entries", () => {
	const { container } = render(
		<TrendingSection type="game" list={{ computedAt: null, items: [] }} />,
	);
	expect(container.innerHTML).toBe("");
});

test("renders nothing when the trending prop is absent", () => {
	const { container } = render(<TrendingSection type="game" />);
	expect(container.innerHTML).toBe("");
});

test("links carry from=trending and prior entries get the Popular chip", () => {
	render(
		<TrendingSection
			type="game"
			list={{ computedAt: "x", items: [entry(1, "trend"), entry(2, "prior")] }}
		/>,
	);
	expect(screen.getByRole("heading").textContent).toContain("home.trending.title");
	expect(screen.getByRole("link", { name: "Game 1" }).getAttribute("href")).toBe(
		"/games/game-1?from=trending",
	);
	expect(screen.getAllByText("home.trending.popular")).toHaveLength(1);
});

test("all-prior list becomes Popular {type} without chips", () => {
	render(
		<TrendingSection
			type="game"
			list={{ computedAt: "x", items: [entry(1, "prior"), entry(2, "prior")] }}
		/>,
	);
	expect(screen.getByRole("heading").textContent).toContain("home.trending.popularType");
	expect(screen.queryByText("home.trending.popular")).toBeNull();
});
