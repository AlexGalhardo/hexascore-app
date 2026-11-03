import { afterEach, beforeEach, expect, test, vi } from "vitest";
import { cleanup, render, screen, within } from "../../lib/testing";
import Offers, { type OffersData } from "./Offers";

vi.mock("../../lib/useT", () => ({
	useT: () => (key: string, p?: Record<string, string | number>) =>
		p ? `${key} ${JSON.stringify(p)}` : key,
}));

beforeEach(() => {
	window.history.replaceState(null, "", "/movies/the-matrix");
	try {
		localStorage.clear();
	} catch {
		/* no storage */
	}
});
afterEach(cleanup);

const data = (over: Partial<OffersData> = {}): OffersData => ({
	countries: ["BR", "US", "PT", "ES", "MX", "AR"],
	defaultCountry: "BR",
	byCountry: {
		BR: [
			{
				id: "o1",
				name: "Netflix",
				kind: "stream",
				verifiedAt: "2026-10-30",
				sponsored: false,
			},
			{ id: "o2", name: "Apple TV", kind: "buy", sponsored: false },
		],
		US: [{ id: "o3", name: "Tubi TV", kind: "free", sponsored: false }],
		PT: [],
		ES: [],
		MX: [],
		AR: [],
	},
	commission: false,
	amazonAssociate: false,
	attribution: ["JustWatch"],
	...over,
});

test("groups the links by kind and sends each through /go with the sponsored rel", () => {
	render(<Offers offers={data()} type="movie" title="The Matrix" locale="en" />);

	const stream = screen.getByRole("group", { name: "offers.kinds.stream" });
	const link = within(stream).getByRole("link", { name: /Netflix/ });
	expect(link.getAttribute("href")).toBe("/go/o1");
	expect(link.getAttribute("target")).toBe("_blank");
	expect(link.getAttribute("rel")).toBe("sponsored nofollow noopener noreferrer");
	expect(link.getAttribute("title")).toBe('offers.verified {"date":"2026-10-30"}');

	expect(
		within(screen.getByRole("group", { name: "offers.kinds.buy" })).getByRole("link", {
			name: /Apple TV/,
		}),
	).toBeTruthy();
});

test("the country switch changes the list and keeps the choice in the URL", () => {
	render(<Offers offers={data()} type="movie" title="The Matrix" locale="en" />);

	const select = screen.getByRole("combobox", { name: "offers.country" }) as HTMLSelectElement;
	expect(select.value).toBe("BR");

	select.value = "US";
	select.dispatchEvent(new Event("change", { bubbles: true }));

	expect(screen.getByRole("link", { name: /Tubi TV/ })).toBeTruthy();
	expect(screen.queryByRole("link", { name: /Netflix/ })).toBeNull();
	expect(window.location.search).toBe("?country=US");
});

test("a country in the URL wins over the locale default", () => {
	window.history.replaceState(null, "", "/movies/the-matrix?country=US");
	render(<Offers offers={data()} type="movie" title="The Matrix" locale="en" />);

	expect((screen.getByRole("combobox") as HTMLSelectElement).value).toBe("US");
	expect(screen.getByRole("link", { name: /Tubi TV/ })).toBeTruthy();
});

test("an unknown country in the URL is ignored", () => {
	window.history.replaceState(null, "", "/movies/the-matrix?country=XX");
	render(<Offers offers={data()} type="movie" title="The Matrix" locale="en" />);

	expect((screen.getByRole("combobox") as HTMLSelectElement).value).toBe("BR");
});

test("with nothing known it says so and offers the JustWatch search for films and series", () => {
	const empty = data({ defaultCountry: "PT", attribution: [] });
	render(<Offers offers={empty} type="series" title="Dark & Co" locale="pt-BR" />);

	expect(screen.getByText(/offers.none/)).toBeTruthy();

	const search = screen.getByRole("link", { name: "offers.searchJustWatch" });
	expect(search.getAttribute("href")).toBe(
		"https://www.justwatch.com/pt/search?q=Dark%20%26%20Co",
	);
	expect(search.getAttribute("rel")).toBe("noopener noreferrer nofollow");
});

test("games and books get no JustWatch search", () => {
	render(
		<Offers offers={data({ defaultCountry: "PT" })} type="game" title="Portal" locale="en" />,
	);

	expect(screen.getByText(/offers.none/)).toBeTruthy();
	expect(screen.queryByRole("link", { name: "offers.searchJustWatch" })).toBeNull();
});

test("the commission and Amazon lines appear only when asked for, the credit when a source is listed", () => {
	const { rerender } = render(
		<Offers offers={data()} type="movie" title="The Matrix" locale="en" />,
	);
	expect(screen.queryByText("offers.commission")).toBeNull();
	expect(screen.queryByText("offers.amazon")).toBeNull();
	expect(screen.getByText("offers.justwatch")).toBeTruthy();

	rerender(
		<Offers
			offers={data({ commission: true, amazonAssociate: true, attribution: [] })}
			type="book"
			title="Dune"
			locale="en"
		/>,
	);
	expect(screen.getByText("offers.commission")).toBeTruthy();
	expect(screen.getByText("offers.amazon")).toBeTruthy();
	expect(screen.queryByText("offers.justwatch")).toBeNull();
});
