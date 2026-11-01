import { cleanup, render, screen, within } from "@testing-library/react";
import { afterEach, expect, test, vi } from "vitest";

import CoverageTable from "./CoverageTable";

vi.mock("../../lib/useT", () => ({
	useT: () => (key: string, params?: Record<string, string | number>) =>
		params ? `${key} ${JSON.stringify(params)}` : key,
}));
vi.mock("@inertiajs/react", () => ({ usePage: () => ({ props: { locale: "en" } }) }));

afterEach(cleanup);

test("renders one row per field and one column per type with the measured percentages", () => {
	render(
		<CoverageTable
			coverage={{
				measuredAt: "2026-11-01",
				rows: [
					{ type: "book", field: "title", items: 4, pct: 100 },
					{ type: "game", field: "title", items: 2, pct: 50 },
					{ type: "book", field: "synopsis", items: 4, pct: 33.3 },
				],
			}}
		/>,
	);
	const table = screen.getByRole("table");
	expect(within(table).getByText(/coverage\.caption.*2026-11-01/)).toBeTruthy();
	expect(within(table).getByRole("columnheader", { name: "catalog.title.book" })).toBeTruthy();
	const title = within(table).getByRole("row", { name: /^title/ });
	expect(within(title).getByText("100%")).toBeTruthy();
	expect(within(title).getByText("50%")).toBeTruthy();
	const synopsis = within(table).getByRole("row", { name: /^synopsis/ });
	expect(within(synopsis).getByText("33.3%")).toBeTruthy();
	expect(within(synopsis).getByText("–")).toBeTruthy();
});

test("states that the numbers are not measured yet instead of an empty table", () => {
	render(<CoverageTable coverage={{ measuredAt: null, rows: [] }} />);
	expect(screen.queryByRole("table")).toBeNull();
	expect(screen.getByText("coverage.notMeasured")).toBeTruthy();
});
