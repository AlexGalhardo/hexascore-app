import userEvent from "@testing-library/user-event";
import { afterEach, expect, test, vi } from "vitest";
import { cleanup, fireEvent, render, screen } from "../../lib/testing";

import EditionPicker from "./EditionPicker";

const put = vi.fn();
vi.mock("../../lib/useLocalized", () => ({ useLocalized: () => (path: string) => path }));
vi.mock("../../lib/useT", () => ({ useT: () => (key: string) => key }));
vi.mock("@inertiajs/react", () => ({ router: { put: (...a: unknown[]) => put(...a) } }));

afterEach(() => {
	cleanup();
	vi.unstubAllGlobals();
});

const stub = (res: Partial<Response>) => vi.stubGlobal("fetch", vi.fn().mockResolvedValue(res));

test("loads the editions and saves the chosen one", async () => {
	stub({
		ok: true,
		json: async () => ({
			data: [
				{ id: "e1", title: "One", language: "en", publisher: null, publicationYear: 1999 },
				{ id: "e2", title: "Two", language: "pt", publisher: "P", publicationYear: null },
			],
		}),
	});
	const user = userEvent.setup();
	render(<EditionPicker itemId="i1" />);
	await user.click(screen.getByRole("button", { name: "library.chooseEdition" }));
	await user.selectOptions(await screen.findByRole("combobox"), "e2");
	await user.click(screen.getByRole("button", { name: "library.saveEdition" }));
	expect(put).toHaveBeenCalledWith(
		"/items/i1/edition",
		{ editionId: "e2" },
		{ preserveScroll: true },
	);
});

test("shows an error when the list fails to load", async () => {
	stub({ ok: false });
	render(<EditionPicker itemId="i1" />);
	await userEvent.setup().click(screen.getByRole("button", { name: "library.chooseEdition" }));
	expect((await screen.findByRole("alert")).textContent).toBe("library.editionsError");
});

test("a second click while a page loads does not fetch it twice", async () => {
	let resolve: (r: Partial<Response>) => void = () => {};
	const fetchMock = vi.fn(() => new Promise<Partial<Response>>((r) => (resolve = r)));
	vi.stubGlobal("fetch", fetchMock);
	render(<EditionPicker itemId="i1" />);
	const button = screen.getByRole("button", { name: "library.chooseEdition" });
	fireEvent.click(button);
	fireEvent.click(button);
	resolve({ ok: true, json: async () => ({ data: [], next: -1 }) });
	await screen.findByText("library.noEditions");
	expect(fetchMock).toHaveBeenCalledTimes(1);
});

test("appends the next page of editions and keeps the first choice", async () => {
	const fetchMock = vi.fn((url: string) =>
		Promise.resolve({
			ok: true,
			json: async () =>
				url.endsWith("cursor=0")
					? {
							data: [
								{
									id: "e1",
									title: "One",
									language: "en",
									publisher: null,
									publicationYear: null,
								},
							],
							next: 5,
						}
					: {
							data: [
								{
									id: "e2",
									title: "Two",
									language: "pt",
									publisher: null,
									publicationYear: null,
								},
							],
							next: -1,
						},
		}),
	);
	vi.stubGlobal("fetch", fetchMock);
	const user = userEvent.setup();
	render(<EditionPicker itemId="i1" />);
	await user.click(screen.getByRole("button", { name: "library.chooseEdition" }));
	await user.click(await screen.findByRole("button", { name: "library.moreEditions" }));
	await screen.findByRole("option", { name: "Two · pt" });
	expect(String(fetchMock.mock.calls[1][0])).toBe("/items/i1/editions?cursor=5");
	expect(screen.queryByRole("button", { name: "library.moreEditions" })).toBeNull();
	expect((screen.getByRole("combobox") as HTMLSelectElement).value).toBe("e1");
});
