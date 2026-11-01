import { render, screen } from "@testing-library/react";
import type { ReactNode } from "react";
import { expect, test, vi } from "vitest";

import Edition from "./Edition";

vi.mock("../../lib/useT", () => ({ useT: () => (key: string) => key }));
vi.mock("../../components/Seo", () => ({ default: () => null }));
vi.mock("@inertiajs/react", () => ({
	Link: ({ href, children }: { href: string; children: ReactNode }) => (
		<a href={href}>{children}</a>
	),
	usePage: () => ({ props: { locale: "en" } }),
}));

test("edition shows its facts, skips missing ones and links back to the book", () => {
	render(
		<Edition
			work={{ id: "w1", slug: "neuromancer", title: "Neuromancer" }}
			edition={{
				id: "e1",
				title: "Neuromancer",
				subtitle: null,
				language: "English",
				publisher: null,
				publicationYear: 1984,
				pageCount: 271,
				format: null,
				isbn13: "9780441569595",
			}}
		/>,
	);
	expect(screen.getByText("271")).toBeTruthy();
	expect(screen.queryByText("item.editionFields.publisher")).toBeNull();
	expect(screen.getByRole("link", { name: "Neuromancer" }).getAttribute("href")).toBe(
		"/books/neuromancer",
	);
	expect(screen.getByText("item.editions.backToWork").closest("a")?.getAttribute("href")).toBe(
		"/books/neuromancer",
	);
});
