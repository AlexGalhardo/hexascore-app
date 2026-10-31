import { render, screen } from "@testing-library/react";
import type { ReactNode } from "react";
import { expect, test, vi } from "vitest";

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

const item = {
	id: "i1",
	type: "game" as const,
	slug: "zelda",
	title: "Zelda",
	year: 2017,
	cover: { thumb: "a/96.webp", full: "a/600.webp" },
};

test("renders cover, status label and score", () => {
	const { container } = render(<ShelfCard item={item} statusLabel="Playing" stars={8.5} />);
	expect(container.querySelector("img")).toBeTruthy();
	expect(screen.getByText("Zelda")).toBeTruthy();
	expect(screen.getByText("Playing")).toBeTruthy();
	expect(screen.getByText("library.score:8.5")).toBeTruthy();
});

test("missing score renders the dash", () => {
	render(<ShelfCard item={item} statusLabel="Playing" stars={null} />);
	expect(screen.getByText("—")).toBeTruthy();
});
