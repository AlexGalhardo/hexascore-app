import { cleanup, render, screen, within } from "@testing-library/react";
import type { ReactNode } from "react";
import { afterEach, expect, test, vi } from "vitest";

import MoreLikeThis from "./MoreLikeThis";

vi.mock("../../lib/useT", () => ({
	useT: () => (key: string, p?: Record<string, string>) =>
		p?.entity ? `${key}:${p.entity}` : key,
}));
vi.mock("@inertiajs/react", () => ({
	Link: ({ children, ...rest }: { href: string; children: ReactNode }) => (
		<a {...rest}>{children}</a>
	),
	usePage: () => ({ props: { locale: "pt-BR", coversBase: "" } }),
}));

afterEach(cleanup);

const hl = (id: string, type: string, score: number | null) => ({
	id,
	type,
	slug: `slug-${id}`,
	title: `Title ${id}`,
	year: 2004,
	score,
	ratingCount: 9,
	cover: null,
});

test("similar cards show reason, year and score and link with from=similar", () => {
	render(
		<MoreLikeThis
			type="game"
			similar={[
				{ ...hl("a", "game", 8.1), reason: { kind: "franchise", entity: "Warcraft" } },
				{ ...hl("b", "game", null), reason: null },
			]}
			crossMedia={[]}
		/>,
	);
	const strip = screen.getByRole("region", { name: "item.moreLike.game" });
	expect(within(strip).getAllByRole("listitem")).toHaveLength(2);
	expect(within(strip).getByRole("link", { name: "Title a" }).getAttribute("href")).toBe(
		"/pt-br/games/slug-a?from=similar",
	);
	expect(within(strip).getByText("item.similarReason:Warcraft")).toBeTruthy();
	expect(within(strip).getByText("8,1")).toBeTruthy();
	expect(screen.queryByRole("region", { name: "item.crossMedia" })).toBeNull();
});

test("cross-media cards name the medium and relation; nothing renders when both lists are empty", () => {
	const { container } = render(<MoreLikeThis type="book" similar={[]} crossMedia={[]} />);
	expect(container.innerHTML).toBe("");
	render(
		<MoreLikeThis
			type="book"
			similar={[]}
			crossMedia={[{ ...hl("m", "movie", 7), relation: "adapted_as" }]}
		/>,
	);
	const strip = screen.getByRole("region", { name: "item.crossMedia" });
	expect(within(strip).getByText("itemTypes.movie · item.relation.adapted_as")).toBeTruthy();
	expect(within(strip).getByRole("link", { name: "Title m" }).getAttribute("href")).toBe(
		"/pt-br/movies/slug-m?from=crossmedia",
	);
});
