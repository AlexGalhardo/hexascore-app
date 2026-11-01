import { cleanup, render, screen } from "@testing-library/react";
import type { ReactNode } from "react";
import { afterEach, expect, test, vi } from "vitest";

import ShelfCard from "./ShelfCard";

vi.mock("../../lib/useT", () => ({
	useT: () => (key: string, p?: Record<string, unknown>) => (p ? `${key}:${p.stars}` : key),
}));
vi.mock("@inertiajs/react", () => ({
	Link: ({ href, children }: { href: string; children: ReactNode }) => (
		<a href={href}>{children}</a>
	),
	usePage: () => ({ props: { locale: "en", coversBase: "/covers" } }),
}));

afterEach(cleanup);

const item = {
	id: "i1",
	type: "game" as const,
	slug: "zelda",
	title: "Zelda",
	year: 2017,
	cover: { thumb: "a/96.webp", full: "a/600.webp" },
};

test("renders cover, status label and score", () => {
	const { container } = render(<ShelfCard item={item} statusLabel="Playing" stars={8.5} own />);
	expect(container.querySelector("img")).toBeTruthy();
	expect(screen.getByText("Zelda")).toBeTruthy();
	expect(screen.getByText("itemTypes.game · 2017 · Playing")).toBeTruthy();
	expect(screen.getByText("library.yourScore:8.5")).toBeTruthy();
});

test("no own score reads not rated by you", () => {
	render(<ShelfCard item={item} statusLabel="Playing" stars={null} own />);
	expect(screen.getByText("library.notRated")).toBeTruthy();
});

test("someone else's score never reads as the viewer's", () => {
	render(<ShelfCard item={item} stars={7} />);
	expect(screen.getByText("library.ratedScore:7.0")).toBeTruthy();
	expect(screen.queryByText(/library\.yourScore/)).toBeNull();
});

test("someone else's unrated item does not say not rated by you", () => {
	render(<ShelfCard item={item} stars={null} />);
	expect(screen.queryByText("library.notRated")).toBeNull();
});
