import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, expect, test, vi } from "vitest";

import EditionPicker from "./EditionPicker";

const put = vi.fn();
vi.mock("../../lib/useT", () => ({ useT: () => (key: string) => key }));
vi.mock("@inertiajs/react", () => ({ router: { put: (...a: unknown[]) => put(...a) } }));

afterEach(() => vi.unstubAllGlobals());

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
