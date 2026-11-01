import { cleanup, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, expect, test, vi } from "vitest";

import CommentThread, { type CommentEntry } from "./CommentThread";
import ReactionBar from "./ReactionBar";

const post = vi.fn();
const put = vi.fn();
const del = vi.fn();

vi.mock("@inertiajs/react", () => ({
	useForm: (data: Record<string, unknown>) => ({
		data,
		setData: vi.fn(),
		post: vi.fn(),
		reset: vi.fn(),
		processing: false,
		errors: {},
	}),
	router: {
		post: (...args: unknown[]) => post(...args),
		put: (...args: unknown[]) => put(...args),
		delete: (...args: unknown[]) => del(...args),
	},
}));
vi.mock("../../lib/useLocalized", () => ({ useLocalized: () => (path: string) => path }));
vi.mock("../../lib/useT", () => ({
	useT: () => (key: string, params?: Record<string, unknown>) =>
		params && "count" in params ? `${key} ${params.count}` : key,
}));

afterEach(() => {
	cleanup();
	post.mockClear();
	put.mockClear();
	del.mockClear();
});

const c = (over: Partial<CommentEntry>): CommentEntry => ({
	id: "c1",
	author: { nickname: "Second", initial: "S" },
	body: "<b>not html</b>",
	likes: 0,
	helpful: 0,
	createdAt: "2026-11-01T10:00:00Z",
	...over,
});

test("renders a thread as text, replies to a top-level comment and deletes your own", async () => {
	render(
		<CommentThread
			target="review"
			id="r1"
			canWrite
			comments={[c({ replies: [c({ id: "c2", body: "reply", mine: true })] })]}
		/>,
	);

	expect(screen.getByText("<b>not html</b>")).toBeTruthy();
	expect(screen.getAllByRole("button", { name: "comments.reply" })).toHaveLength(1);

	await userEvent.click(screen.getByRole("button", { name: "comments.reply" }));
	await userEvent.type(screen.getByRole("textbox", { name: "comments.replyLabel" }), "me too");
	// The reply form comes first in the DOM, above the top-level form.
	await userEvent.click(screen.getAllByRole("button", { name: "comments.send" })[0]);
	expect(post).toHaveBeenCalledWith(
		"/comments/review/r1",
		{ body: "me too", parentId: "c1" },
		expect.anything(),
	);

	await userEvent.click(screen.getByRole("button", { name: "comments.delete" }));
	expect(del).toHaveBeenCalledWith("/comments/c2", expect.anything());
});

test("posts a top-level comment", async () => {
	render(<CommentThread target="list" id="l1" canWrite comments={[]} />);

	await userEvent.type(screen.getByRole("textbox", { name: "comments.label" }), "Nice list");
	await userEvent.click(screen.getByRole("button", { name: "comments.send" }));
	expect(post).toHaveBeenCalledWith(
		"/comments/list/l1",
		{ body: "Nice list", parentId: "" },
		expect.anything(),
	);
});

test("reaction buttons toggle one reaction per target", async () => {
	render(<ReactionBar target="review" id="r1" likes={2} helpful={1} mine="like" canReact />);

	const like = screen.getByRole("button", { name: "reactions.like 2" });
	expect(like.getAttribute("aria-pressed")).toBe("true");

	await userEvent.click(like);
	expect(del).toHaveBeenCalledWith("/reactions/review/r1", expect.anything());

	await userEvent.click(screen.getByRole("button", { name: "reactions.helpful 1" }));
	expect(put).toHaveBeenCalledWith(
		"/reactions/review/r1",
		{ kind: "helpful" },
		expect.anything(),
	);
});

test("reactions are read-only without a signed-in reader", () => {
	render(<ReactionBar target="list" id="l1" likes={3} helpful={0} canReact={false} />);

	expect(screen.queryByRole("button")).toBeNull();
	expect(screen.getByText("reactions.like 3")).toBeTruthy();
});

test("offers a report button on other people's comments only", () => {
	render(
		<CommentThread
			target="review"
			id="r1"
			canWrite
			comments={[c({ replies: [c({ id: "c2", body: "mine", mine: true })] })]}
		/>,
	);

	expect(screen.getAllByRole("button", { name: "social.report.reportComment" })).toHaveLength(1);
});
