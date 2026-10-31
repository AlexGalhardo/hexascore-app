import { cleanup, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, expect, test, vi } from "vitest";

import SuggestEdit from "./SuggestEdit";

const patch = vi.fn();
const post = vi.fn();

vi.mock("../../lib/useT", () => ({
	useT: () => (key: string, p?: Record<string, string | number>) =>
		p ? `${key} ${JSON.stringify(p)}` : key,
}));
vi.mock("../../components/Seo", () => ({ default: () => null }));
vi.mock("@inertiajs/react", () => ({
	router: { patch: (...a: unknown[]) => patch(...a), post: (...a: unknown[]) => post(...a) },
	usePage: () => ({ props: { errors: {} } }),
}));

const target = {
	kind: "item",
	id: "i1",
	domain: "movie",
	values: { title: "Old", "translations.en.title": null },
};

beforeEach(() => {
	cleanup();
	patch.mockClear();
	post.mockClear();
});

test("closed flag shows the notice only", () => {
	render(<SuggestEdit open={false} />);
	expect(screen.getByText("community.closed.title")).toBeTruthy();
});

test("shows quota, restores the draft and disables a locked field with its reason", () => {
	render(
		<SuggestEdit
			open
			target={target}
			fields={["title", "translations.en.title"]}
			locks={[{ field: "translations.en.title", reason: "Disputed" }]}
			quota={{ limit: 3, used: 1, left: 2 }}
			draft={{
				id: "d1",
				version: 2,
				locale: "en",
				note: "",
				sourceUrls: ["https://www.wikidata.org/wiki/Q1"],
				changes: { title: { after: "Draft title" } },
			}}
		/>,
	);
	expect(screen.getByTestId("quota-left").textContent).toContain('"left":2');
	expect((screen.getByDisplayValue("Draft title") as HTMLInputElement).value).toBe("Draft title");
	expect(screen.getByText(/Disputed/)).toBeTruthy();
	const locked = screen.getAllByRole("textbox").find((e) => (e as HTMLInputElement).disabled);
	expect(locked).toBeTruthy();
});

test("saving an existing draft patches it with its version and only changed fields", async () => {
	const user = userEvent.setup();
	render(
		<SuggestEdit
			open
			target={target}
			fields={["title"]}
			draft={{ id: "d1", version: 2, locale: "en", note: "", sourceUrls: [], changes: {} }}
		/>,
	);
	const input = screen.getByDisplayValue("Old");
	await user.clear(input);
	await user.type(input, "New");
	await user.click(screen.getByRole("button", { name: "community.saveDraft" }));
	expect(patch).toHaveBeenCalledWith(
		"/suggestions/d1",
		expect.objectContaining({
			version: 2,
			website: "",
			changes: { title: { before: "Old", after: "New" } },
		}),
		expect.anything(),
	);
});
