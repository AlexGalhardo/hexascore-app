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
	expect((screen.getByDisplayValue("src-1") as HTMLInputElement).value).toBe("src-1");
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
