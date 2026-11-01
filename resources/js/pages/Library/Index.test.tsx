import userEvent from "@testing-library/user-event";
import type { ReactNode } from "react";
import { expect, test, vi } from "vitest";
import { render, screen } from "../../lib/testing";

import Index from "./Index";

const del = vi.fn();
vi.mock("../../lib/useT", () => ({
	useT: () => (key: string, p?: { count?: number }) =>
		p?.count === undefined ? key : `${key}:${p.count}`,
}));
vi.mock("../../components/Seo", () => ({ default: () => null }));
vi.mock("../../components/library/ItemLink", () => ({ default: () => null }));
vi.mock("@inertiajs/react", () => ({
	Link: ({ href, children }: { href: string; children: ReactNode }) => (
		<a href={href}>{children}</a>
	),
	router: { get: vi.fn(), delete: (...a: unknown[]) => del(...a) },
	usePage: () => ({ props: { locale: "en" } }),
}));

const props = {
	counts: { want: { game: 2, book: 1 }, completed: { game: 3 } },
	groups: {
		data: [
			{
				type: "game" as const,
				total: 5,
				next: 30,
				items: [
					{
						id: "i1",
						type: "game" as const,
						slug: "s",
						title: "T",
						year: null,
						coverThumbKey: null,
						status: "want",
						stars: null,
						edition: { id: "e1", title: "Ed" },
					},
				],
			},
		],
	},
};

test("shows filter counts and a see-all link per truncated group", () => {
	render(<Index {...props} filters={{ status: "", type: "", sort: "" }} />);
	expect(screen.getByRole("option", { name: /shelves\.want.*\(3\)/ })).toBeTruthy();
	expect(screen.getByRole("option", { name: /itemTypes\.game.*\(5\)/ })).toBeTruthy();
	expect(screen.getByText("library.seeAll").closest("a")?.getAttribute("href")).toBe(
		"/library?type=game",
	);
});

test("clear edition deletes the item edition", async () => {
	render(<Index {...props} filters={{ status: "", type: "", sort: "" }} />);
	await userEvent
		.setup()
		.click(screen.getAllByRole("button", { name: "library.clearEdition" })[0]);
	expect(del).toHaveBeenCalledWith("/items/i1/edition", { preserveScroll: true });
});

test("groups cards under a heading per status", () => {
	const { container } = render(<Index {...props} filters={{ status: "", type: "", sort: "" }} />);
	expect(container.querySelector("#status-want")?.textContent).toContain("library.titleCount:3");
	expect(container.querySelector("#status-completed")).toBeNull();
});
