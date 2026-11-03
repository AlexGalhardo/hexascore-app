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
	render(<Offers offers={data()} type="movie" />);

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

test("offers from every country show for any visitor, one link per store and kind", () => {
	render(<Offers offers={data()} type="movie" />);
	expect(screen.queryByRole("combobox")).toBeNull();
	expect(screen.getByRole("link", { name: /Tubi TV/ }).getAttribute("href")).toBe("/go/o3");
	expect(screen.getAllByRole("link", { name: /Netflix/ })).toHaveLength(1);
});

test("with no offers anywhere it renders nothing, for films, series and games", () => {
	const empty = data({ byCountry: {} });
	for (const type of ["movie", "series", "game"] as const) {
		const { container, unmount } = render(<Offers offers={empty} type={type} />);
		expect(container.innerHTML).toBe("");
		unmount();
	}
});

test("the commission and Amazon lines appear only when asked for, the credit when a source is listed", () => {
	const { rerender } = render(<Offers offers={data()} type="movie" />);
	expect(screen.queryByText("offers.commission")).toBeNull();
	expect(screen.queryByText("offers.amazon")).toBeNull();
	expect(screen.getByText("offers.justwatch")).toBeTruthy();

	rerender(
		<Offers
			offers={data({ commission: true, amazonAssociate: true, attribution: [] })}
			type="book"
		/>,
	);
	expect(screen.getByText("offers.commission")).toBeTruthy();
	expect(screen.getByText("offers.amazon")).toBeTruthy();
	expect(screen.queryByText("offers.justwatch")).toBeNull();
});
