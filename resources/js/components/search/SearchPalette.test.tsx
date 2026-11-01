import { cleanup, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { createRef } from "react";
import { afterEach, beforeEach, expect, test, vi } from "vitest";

import SearchPalette from "./SearchPalette";

const visit = vi.fn();

vi.mock("@inertiajs/react", () => ({
	router: { visit: (...a: unknown[]) => visit(...a) },
	usePage: () => ({ props: { locale: "en" } }),
}));
vi.mock("../../lib/useT", () => ({
	useT: () => (key: string, p?: Record<string, string>) => (p?.q ? `${key}:${p.q}` : key),
}));

const data = Array.from({ length: 10 }, (_, i) => ({
	id: String(i),
	type: "game",
	kind: "game",
	title: `Zelda ${i}`,
	year: 1986 + i,
	path: `/games/zelda-${i}`,
}));
let answer = data;

beforeEach(() => {
	visit.mockReset();
	localStorage.clear();
	answer = data;
	vi.stubGlobal(
		"fetch",
		vi.fn(async () => ({ ok: true, json: async () => ({ data: answer }) })),
	);
});
afterEach(() => {
	cleanup();
	vi.unstubAllGlobals();
});

function open() {
	render(<SearchPalette open onOpenChange={() => {}} returnFocus={createRef()} />);
	return screen.getByRole("combobox");
}

test("focuses the input and shows the first 8 results plus see all", async () => {
	const box = open();
	await waitFor(() => expect(document.activeElement).toBe(box));
	await userEvent.type(box, "zel");
	const options = await screen.findAllByRole("option");
	expect(options).toHaveLength(9);
	expect(options[8].textContent).toBe("search.palette.seeAll:zel");
	expect(options[0].getAttribute("aria-selected")).toBe("true");
});

test("arrow keys move the highlight and Enter opens the item, remembering the query", async () => {
	const box = open();
	await userEvent.type(box, "zel");
	await screen.findAllByRole("option");
	await userEvent.keyboard("{ArrowDown}{ArrowDown}{ArrowUp}{Enter}");
	expect(visit).toHaveBeenCalledWith("/games/zelda-1");
	expect(JSON.parse(localStorage.getItem("recentSearches") ?? "[]")).toEqual(["zel"]);
});

test("ArrowUp from the first row wraps to see all", async () => {
	const box = open();
	await userEvent.type(box, "zel");
	await screen.findAllByRole("option");
	await userEvent.keyboard("{ArrowUp}{Enter}");
	expect(visit).toHaveBeenCalledWith("/search?q=zel");
});

test("no results shows the message and keeps see all", async () => {
	answer = [];
	const box = open();
	await userEvent.type(box, "qqx");
	expect(await screen.findByText("search.noResults:qqx")).toBeTruthy();
	expect(screen.getAllByRole("option")).toHaveLength(1);
});

test("empty query lists recent searches", async () => {
	localStorage.setItem("recentSearches", JSON.stringify(["witcher", "dune"]));
	const box = open();
	await waitFor(() => expect(document.activeElement).toBe(box));
	const options = await screen.findAllByRole("option");
	expect(options.map((o) => o.textContent)).toEqual(["witcher", "dune"]);
	await userEvent.keyboard("{Enter}");
	expect(visit).toHaveBeenCalledWith("/search?q=witcher");
});
