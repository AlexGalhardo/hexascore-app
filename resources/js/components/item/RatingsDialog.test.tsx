import userEvent from "@testing-library/user-event";
import { afterEach, expect, test, vi } from "vitest";

import { cleanup, render, screen, waitFor } from "../../lib/testing";
import RatingsDialog from "./RatingsDialog";

vi.mock("../../lib/useLocalized", () => ({ useLocalized: () => (path: string) => path }));
vi.mock("../../lib/useT", () => ({ useT: () => (key: string) => key }));
// jsdom has no layout, so no scrollTo.
Element.prototype.scrollTo = () => {};

afterEach(() => {
	cleanup();
	vi.unstubAllGlobals();
});

const page = (n: number, count: number) => ({
	ok: true,
	json: async () => ({
		data: [{ nickname: `user${n}`, overall: 8, categories: [7], ratedAt: `2026-0${n}` }],
		page: n,
		page_count: count,
	}),
});

test("fetches nothing while closed", () => {
	const fetchMock = vi.fn();
	vi.stubGlobal("fetch", fetchMock);
	render(<RatingsDialog itemId="i1" categoryLabels={["Story"]} />);
	expect(fetchMock).not.toHaveBeenCalled();
});

test("loads the first page on open and the next page on demand", async () => {
	const fetchMock = vi.fn((url: string) =>
		Promise.resolve(page(url.includes("page=2") ? 2 : 1, 2)),
	);
	vi.stubGlobal("fetch", fetchMock);
	render(<RatingsDialog itemId="i1" categoryLabels={["Story"]} initialOpen />);
	expect(await screen.findByText("user1")).toBeTruthy();
	expect(String(fetchMock.mock.calls[0][0])).toBe("/api/items/i1/ratings?page=1");
	await userEvent.click(screen.getByRole("button", { name: "item.next" }));
	expect(await screen.findByText("user2")).toBeTruthy();
	expect(fetchMock).toHaveBeenCalledTimes(2);
});

test("shows the error and retries on demand", async () => {
	const fetchMock = vi
		.fn()
		.mockResolvedValueOnce({ ok: false, status: 500 })
		.mockResolvedValueOnce(page(1, 1));
	vi.stubGlobal("fetch", fetchMock);
	render(<RatingsDialog itemId="i1" categoryLabels={["Story"]} initialOpen />);
	await screen.findByRole("alert");
	await userEvent.click(screen.getByRole("button", { name: "item.retry" }));
	await waitFor(() => expect(screen.queryByRole("alert")).toBeNull());
	expect(await screen.findByText("user1")).toBeTruthy();
	expect(fetchMock).toHaveBeenCalledTimes(2);
});

test("the users' scores modal is wide, with larger type and full category names", async () => {
	vi.stubGlobal(
		"fetch",
		vi.fn(() => Promise.resolve(page(1, 1))),
	);
	render(<RatingsDialog itemId="i1" categoryLabels={["Cinematography"]} initialOpen />);
	await screen.findByText("user1");

	const dialog = screen.getByRole("dialog");
	expect(dialog.className).toContain("sm:max-w-5xl");
	expect(screen.getByRole("columnheader", { name: "Cinematography" }).textContent).toBe(
		"Cinematography",
	);
	expect(screen.getByRole("table").className).toContain("text-xl");
});
