import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { afterEach, expect, test, vi } from "vitest";

import Filters from "./Filters";

vi.mock("../../lib/useT", () => ({ useT: () => (key: string) => key }));

afterEach(cleanup);

const filters = { genre: "", platform: "", year: 0, minScore: null, sort: "-popularity", per: 50 };

test("the page size is a GET field with the four sizes the server honours", () => {
	render(<Filters action="/games" type="game" filters={filters} genres={[]} platforms={[]} />);
	const per = screen.getByRole("combobox", {
		name: "catalog.perPage",
		hidden: true,
	}) as HTMLSelectElement;
	expect(per.name).toBe("per");
	expect(per.value).toBe("50");
	expect([...per.options].map((o) => o.value)).toEqual(["25", "50", "75", "100"]);
});

test("below lg the fields sit behind a disclosure button", () => {
	render(<Filters action="/games" type="game" filters={filters} genres={[]} platforms={[]} />);
	const toggle = screen.getByRole("button", { name: "catalog.filters" });
	expect(toggle.getAttribute("aria-expanded")).toBe("false");
	fireEvent.click(toggle);
	expect(toggle.getAttribute("aria-expanded")).toBe("true");
	expect(document.getElementById(toggle.getAttribute("aria-controls") ?? "")?.className).toMatch(
		/^grid /,
	);
});
