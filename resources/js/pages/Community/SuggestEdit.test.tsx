import { cleanup, fireEvent, render, screen } from "@testing-library/react";
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

test("release date and relation edits are sent as typed changes", async () => {
	const user = userEvent.setup();
	render(
		<SuggestEdit
			open
			target={{
				...target,
				values: { title: "Old", release_date: { value: "2001-12-19", precision: "day" } },
			}}
			fields={["release_date", "genres", "creators"]}
			draft={{ id: "d1", version: 1, locale: "en", note: "", sourceUrls: [], changes: {} }}
		/>,
	);
	const date = screen.getByLabelText("community.fields.release_date") as HTMLInputElement;
	expect(date.value).toBe("2001-12-19");
	fireEvent.change(date, { target: { value: "2001-12-18" } });
	await user.type(
		screen.getByLabelText("community.relation.add community.fields.genres"),
		"g1\ng2",
	);
	await user.type(
		screen.getByLabelText("community.relation.remove community.fields.genres"),
		"g3",
	);
	await user.type(
		screen.getByLabelText("community.relation.add community.fields.creators"),
		"p1",
	);
	await user.selectOptions(
		screen.getByLabelText("community.relation.role community.fields.creators"),
		"cast",
	);
	await user.click(screen.getByRole("button", { name: "community.saveDraft" }));
	expect(patch).toHaveBeenCalledWith(
		"/suggestions/d1",
		expect.objectContaining({
			changes: {
				release_date: {
					before: { value: "2001-12-19", precision: "day" },
					after: { value: "2001-12-18", precision: "day" },
				},
				genres: { add: [{ id: "g1" }, { id: "g2" }], remove: [{ id: "g3" }] },
				creators: { add: [{ id: "p1", role: "cast" }], remove: [] },
			},
		}),
		expect.anything(),
	);
});

test("a saved relation draft is restored into the editor", () => {
	render(
		<SuggestEdit
			open
			target={target}
			fields={["genres"]}
			draft={{
				id: "d1",
				version: 1,
				locale: "en",
				note: "",
				sourceUrls: [],
				changes: { genres: { add: [{ id: "g1" }], remove: [{ id: "g9" }] } },
			}}
		/>,
	);
	expect(screen.getByDisplayValue("g1")).toBeTruthy();
	expect(screen.getByDisplayValue("g9")).toBeTruthy();
});
