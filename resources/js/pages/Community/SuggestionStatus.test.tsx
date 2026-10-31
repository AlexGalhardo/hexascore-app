import { cleanup, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { ReactNode } from "react";
import { beforeEach, expect, test, vi } from "vitest";

import MySuggestions from "./MySuggestions";
import SuggestionStatus from "./SuggestionStatus";

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
	router: { post: (...a: unknown[]) => post(...a) },
}));

const base = {
	id: "s1",
	kind: "edit",
	targetType: "item",
	targetId: "i1",
	domain: "movie",
	status: "draft",
	version: 3,
	note: "",
	sourceUrls: [],
	changes: { title: { before: "Old", after: "New" } },
	rejectReason: null,
	decisionNote: null,
	events: [
		{ from: null, to: "draft", actorRole: "author", reason: null, at: "2026-10-31T10:00:00Z" },
	],
};
const quota = { limit: 3, used: 0, left: 3, resetAt: "2026-11-01T00:00:00Z" };

beforeEach(() => {
	cleanup();
	post.mockClear();
});

test("shows the diff with text markers, the timeline and posts submit with the version", async () => {
	render(<SuggestionStatus suggestion={base} canAct quota={quota} open />);
	expect(screen.getByText("Old")).toBeTruthy();
	expect(screen.getByText("New")).toBeTruthy();
	expect(screen.getAllByText("community.status.removed:").length).toBe(1);
	expect(screen.getByText(/community\.status\.states\.draft \(author\)/)).toBeTruthy();
	await userEvent.click(screen.getByRole("button", { name: "community.status.submit" }));
	expect(post).toHaveBeenCalledWith("/suggestions/s1/submit", { version: 3 }, expect.anything());
});

test("a spent quota hides submit and shows the reset time; admins get no actions", () => {
	const { rerender } = render(
		<SuggestionStatus suggestion={base} canAct quota={{ ...quota, left: 0, used: 3 }} open />,
	);
	expect(screen.queryByRole("button", { name: "community.status.submit" })).toBeNull();
	expect(screen.getByRole("status").textContent).toContain("community.status.quotaReached");
	rerender(<SuggestionStatus suggestion={base} canAct={false} quota={quota} open />);
	expect(screen.queryByRole("button", { name: "community.status.withdraw" })).toBeNull();
});

test("rejected shows the reason; applied cannot be withdrawn", () => {
	render(
		<SuggestionStatus
			suggestion={{ ...base, status: "rejected", rejectReason: "no_source" }}
			canAct
			quota={quota}
			open
		/>,
	);
	expect(screen.getByText(/no_source/)).toBeTruthy();
	expect(screen.queryByRole("button", { name: "community.status.withdraw" })).toBeNull();
});

test("my suggestions lists links and an empty state", () => {
	render(
		<MySuggestions
			open
			suggestions={[
				{
					id: "s1",
					kind: "edit",
					domain: "movie",
					status: "submitted",
					title: "Hello",
					updatedAt: "2026-10-31T10:00:00Z",
				},
			]}
		/>,
	);
	expect(screen.getByRole("link", { name: "Hello" }).getAttribute("href")).toBe(
		"/suggestions/s1",
	);
	cleanup();
	render(<MySuggestions open suggestions={[]} />);
	expect(screen.getByText("community.mine.empty")).toBeTruthy();
});
