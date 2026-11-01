import userEvent from "@testing-library/user-event";
import { afterEach, beforeEach, expect, test, vi } from "vitest";
import { cleanup, render, screen, waitFor } from "../../lib/testing";

import SuggestBox from "./SuggestBox";

const visit = vi.fn();

vi.mock("@inertiajs/react", () => ({
	router: { visit: (...a: unknown[]) => visit(...a) },
	usePage: () => ({ props: { locale: "en", messages: {} } }),
}));
vi.mock("../../lib/useT", () => ({ useT: () => (key: string) => key }));

const data = [
	{ id: "1", type: "game", kind: "game", title: "Zelda", year: 1986, path: "/games/zelda" },
	{ id: "2", type: "game", kind: "game", title: "Zelda II", year: 1987, path: "/games/zelda-ii" },
];

beforeEach(() => {
	visit.mockReset();
	vi.stubGlobal(
		"fetch",
		vi.fn(async () => ({ ok: true, json: async () => ({ data }) })),
	);
});
afterEach(() => {
	cleanup();
	vi.unstubAllGlobals();
});

async function typeZel() {
	const user = userEvent.setup();
	const box = screen.getByRole("combobox");
	await user.type(box, "zel");
	await screen.findAllByRole("option");
	return { user, box };
}

test("does not fetch under two characters", async () => {
	const user = userEvent.setup();
	render(<SuggestBox />);
	await user.type(screen.getByRole("combobox"), "z");
	await new Promise((r) => setTimeout(r, 250));
	expect(fetch).not.toHaveBeenCalled();
	expect(screen.getByRole("combobox").getAttribute("aria-expanded")).toBe("false");
});

test("exposes combobox ARIA and moves aria-activedescendant with the arrows", async () => {
	render(<SuggestBox />);
	const { user, box } = await typeZel();
	const options = screen.getAllByRole("option");
	expect(box.getAttribute("aria-expanded")).toBe("true");
	expect(box.getAttribute("aria-controls")).toBe(screen.getByRole("listbox").id);
	expect(box.getAttribute("aria-activedescendant")).toBeNull();

	await user.keyboard("{ArrowDown}");
	expect(box.getAttribute("aria-activedescendant")).toBe(options[0].id);
	expect(options[0].getAttribute("aria-selected")).toBe("true");
	await user.keyboard("{ArrowDown}{ArrowDown}");
	expect(box.getAttribute("aria-activedescendant")).toBe(options[0].id);
	await user.keyboard("{ArrowUp}");
	expect(box.getAttribute("aria-activedescendant")).toBe(options[1].id);
});

test("Enter opens the active option, Escape closes the popup", async () => {
	render(<SuggestBox />);
	const { user, box } = await typeZel();
	await user.keyboard("{ArrowDown}{ArrowDown}{Enter}");
	expect(visit).toHaveBeenCalledWith("/games/zelda-ii");

	await user.type(box, "{Escape}");
	await waitFor(() => expect(box.getAttribute("aria-expanded")).toBe("false"));
});

test("Enter without an active option searches the typed text", async () => {
	render(<SuggestBox />);
	const { user } = await typeZel();
	await user.keyboard("{Enter}");
	expect(visit).toHaveBeenCalledWith("/search?q=zel");
});

test("onPick replaces navigation and the type narrows the query", async () => {
	const onPick = vi.fn();
	render(<SuggestBox onPick={onPick} type="game" label="Add" />);
	const user = userEvent.setup();
	await user.type(screen.getByRole("combobox", { name: "Add" }), "zel");
	const [first] = await screen.findAllByRole("option");
	await user.pointer({ keys: "[MouseLeft]", target: first });
	expect(onPick).toHaveBeenCalledWith(data[0]);
	expect(visit).not.toHaveBeenCalled();
	expect(String(vi.mocked(fetch).mock.calls[0][0])).toContain("&type=game");
});
