import { render, screen } from "@testing-library/react";
import type { ReactNode } from "react";
import { expect, test, vi } from "vitest";

import Show from "./Show";

const seo = vi.fn();
vi.mock("../../lib/useT", () => ({ useT: () => (key: string) => key }));
vi.mock("../../components/Seo", () => ({
	default: (p: unknown) => {
		seo(p);
		return null;
	},
	truncate: (s: string) => s,
}));
vi.mock("../../components/catalog/ItemCard", () => ({
	default: ({ item, href }: { item: { title: string }; href: string }) => (
		<li>
			<a href={href}>{item.title}</a>
		</li>
	),
}));
vi.mock("@inertiajs/react", () => ({
	Link: ({ href, children }: { href: string; children: ReactNode }) => (
		<a href={href}>{children}</a>
	),
	usePage: () => ({ url: "/pt-br/people/ridley-scott?type=movie", props: { locale: "pt-BR" } }),
}));

const props = {
	entity: { id: "e1", kind: "person", slug: "ridley-scott", name: "Ridley Scott" },
	roles: [{ role: "director", count: 2 }],
	items: {
		data: [
			{
				type: "movie" as const,
				total: 14,
				next: 12,
				items: [
					{
						id: "i1",
						type: "movie",
						slug: "alien",
						title: "Alien",
						year: 1979,
						score: null,
						ratingCount: 0,
						cover: null,
					},
				],
			},
		],
	},
	counts: { movie: 14 },
	avgScore: { movie: 7.9 },
	filters: { type: "" },
	indexable: false,
};

test("thin entity is noindex,follow and links items and the next page in the active locale", () => {
	render(<Show {...props} />);
	expect(seo).toHaveBeenCalledWith(expect.objectContaining({ robots: "noindex,follow" }));
	expect(screen.getByText("Alien").getAttribute("href")).toBe("/pt-br/movies/alien");
	expect(screen.getByText("entity.seeAll").closest("a")?.getAttribute("href")).toBe(
		"/pt-br/people/ridley-scott?type=movie",
	);
	expect(screen.getByText(/entity\.avgScore/)).toBeTruthy();
});

test("indexable entity leaves robots to the shared value", () => {
	seo.mockClear();
	render(<Show {...props} indexable />);
	expect(seo).toHaveBeenCalledWith(expect.objectContaining({ robots: undefined }));
});

test("empty entity shows the empty state with a way back to the catalogue", () => {
	render(<Show {...props} items={{ data: [] }} />);
	expect(screen.getByRole("heading", { name: "entity.empty" })).toBeTruthy();
	expect(screen.getByText("entity.browse").closest("a")?.getAttribute("href")).toBe("/pt-br/");
});
