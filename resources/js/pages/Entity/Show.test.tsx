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
	usePage: () => ({
		url: "/pt-br/movies/people/ridley-scott?cursor=12",
		props: { locale: "pt-BR" },
	}),
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
	counts: { movie: 14, book: 3 },
	avgScore: { movie: 7.9 },
	filters: { type: "movie" },
	indexable: false,
};

test("thin entity is noindex,follow and links items, the next page and its other categories", () => {
	render(<Show {...props} />);
	expect(seo).toHaveBeenCalledWith(expect.objectContaining({ robots: "noindex,follow" }));
	expect(screen.getByText("Alien").getAttribute("href")).toBe("/pt-br/movies/alien");
	expect(screen.getByText("entity.next").closest("a")?.getAttribute("href")).toBe(
		"/pt-br/movies/people/ridley-scott?cursor=12",
	);
	const books = screen.getByRole("link", { name: /catalog\.title\.book/ });
	expect(books.getAttribute("href")).toBe("/pt-br/books/people/ridley-scott");
	expect(screen.queryByRole("link", { name: /catalog\.title\.movie/ })).toBeNull();
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
