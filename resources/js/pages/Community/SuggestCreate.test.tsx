import { cleanup, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { ReactNode } from "react";
import { beforeEach, expect, test, vi } from "vitest";

import SuggestCreate from "./SuggestCreate";

const patch = vi.fn();
const post = vi.fn();

vi.mock("../../lib/useT", () => ({
	useT: () => (key: string, p?: Record<string, string | number>) =>
		p ? `${key} ${JSON.stringify(p)}` : key,
}));
vi.mock("../../components/Seo", () => ({ default: () => null }));
vi.mock("../../components/search/SuggestBox", () => ({
	default: ({
		onPick,
		label,
	}: {
		onPick: (s: { id: string; title: string }) => void;
		label: string;
	}) => (
		<button type="button" onClick={() => onPick({ id: `pick-${label}`, title: `T ${label}` })}>
			{label}
		</button>
	),
}));
vi.mock("@inertiajs/react", () => ({
	Link: ({ href, children }: { href: string; children: ReactNode }) => (
		<a href={href}>{children}</a>
	),
	router: { patch: (...a: unknown[]) => patch(...a), post: (...a: unknown[]) => post(...a) },
	usePage: () => ({ props: { errors: {} } }),
}));

beforeEach(() => {
	cleanup();
	patch.mockClear();
	post.mockClear();
});

test("closed flag shows the notice only", () => {
	render(<SuggestCreate open={false} />);
	expect(screen.getByText("community.closed.title")).toBeTruthy();
});

test("create posts the filled fields with no target", async () => {
	render(<SuggestCreate open domain="movie" fields={["title", "synopsis.en"]} />);
	await userEvent.type(screen.getByLabelText(/community\.fields\.title \*/), "New film");
	await userEvent.click(screen.getByRole("button", { name: "community.saveDraft" }));
	expect(post).toHaveBeenCalledOnce();
	const [url, body] = post.mock.calls[0] as [string, Record<string, unknown>];
	expect(url).toBe("/suggestions");
	expect(body).toMatchObject({
		kind: "create",
		domain: "movie",
		targetId: "",
		version: 0,
		changes: { title: { before: null, after: "New film" } },
	});
	expect(Object.keys(body.changes as object)).toEqual(["title"]);
});

test("merge posts the duplicate as target and the kept item as after; a draft patches", async () => {
	render(
		<SuggestCreate
			open
			domain="movie"
			kind="merge"
			draft={{
				id: "d1",
				version: 2,
				locale: "en",
				note: "",
				sourceUrls: [],
				changes: { "*merge*": { before: "src-1", after: "dst-1" } },
			}}
		/>,
	);
	expect(screen.getByText("src-1")).toBeTruthy();
	expect(screen.getByText("dst-1")).toBeTruthy();
	await userEvent.click(screen.getByRole("button", { name: "community.saveDraft" }));
	expect(patch).toHaveBeenCalledWith(
		"/suggestions/d1",
		expect.objectContaining({
			kind: "merge",
			targetId: "src-1",
			version: 2,
			changes: { "*merge*": { before: "src-1", after: "dst-1" } },
		}),
		expect.anything(),
	);
});

test("merge picks both records with the search box and blocks submit until both are set", async () => {
	render(<SuggestCreate open domain="movie" kind="merge" />);
	const submit = screen.getByRole("button", { name: "community.saveDraft" }) as HTMLButtonElement;
	expect(submit.disabled).toBe(true);
	await userEvent.click(screen.getByRole("button", { name: "community.create.mergeSource" }));
	await userEvent.click(screen.getByRole("button", { name: "community.create.mergeDest" }));
	expect(screen.getByText("T community.create.mergeSource")).toBeTruthy();
	expect(submit.disabled).toBe(false);
	await userEvent.click(submit);
	expect(post).toHaveBeenCalledWith(
		"/suggestions",
		expect.objectContaining({
			targetId: "pick-community.create.mergeSource",
			changes: {
				"*merge*": {
					before: "pick-community.create.mergeSource",
					after: "pick-community.create.mergeDest",
				},
			},
		}),
		expect.anything(),
	);
});

test("item creates offer entity and, for books, edition targets", () => {
	const { unmount } = render(<SuggestCreate open domain="game" fields={["title"]} />);
	expect(
		screen.getByRole("link", { name: "community.create.targets.entity" }).getAttribute("href"),
	).toBe("/suggest/new/game?kind=create&target=entity");
	expect(screen.queryByRole("link", { name: "community.create.targets.edition" })).toBeNull();
	unmount();
	render(<SuggestCreate open domain="book" fields={["title"]} />);
	expect(screen.getByRole("link", { name: "community.create.targets.edition" })).toBeTruthy();
});

test("entity create posts its kind and targetKind entity", async () => {
	render(
		<SuggestCreate
			open
			domain="game"
			target="entity"
			fields={["title", "cover", "entity_kind"]}
		/>,
	);
	await userEvent.type(screen.getByLabelText(/community\.fields\.title \*/), "Douglas Adams");
	await userEvent.selectOptions(
		screen.getByLabelText(/community\.fields\.entity_kind/),
		"person",
	);
	await userEvent.click(screen.getByRole("button", { name: "community.saveDraft" }));
	expect(post).toHaveBeenCalledWith(
		"/suggestions",
		expect.objectContaining({
			targetKind: "entity",
			changes: {
				title: { before: null, after: "Douglas Adams" },
				entity_kind: { before: null, after: "person" },
			},
		}),
		expect.anything(),
	);
});

test("edition create picks its book and language and blocks submit until both are set", async () => {
	render(
		<SuggestCreate
			open
			domain="book"
			target="edition"
			fields={["title", "external_ids.isbn13", "work_id", "language"]}
		/>,
	);
	const submit = screen.getByRole("button", { name: "community.saveDraft" }) as HTMLButtonElement;
	expect(submit.disabled).toBe(true);
	await userEvent.click(screen.getByRole("button", { name: "community.fields.work_id *" }));
	await userEvent.type(screen.getByLabelText(/community\.fields\.language/), "pt-BR");
	await userEvent.type(screen.getByLabelText(/community\.fields\.externalId/), "9780000000002");
	expect(submit.disabled).toBe(false);
	await userEvent.click(submit);
	expect(post).toHaveBeenCalledWith(
		"/suggestions",
		expect.objectContaining({
			targetKind: "edition",
			changes: {
				"external_ids.isbn13": { before: null, after: "9780000000002" },
				work_id: { before: null, after: "pick-community.fields.work_id *" },
				language: { before: null, after: "pt-BR" },
			},
		}),
		expect.anything(),
	);
});
