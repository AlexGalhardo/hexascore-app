import { afterEach, beforeEach, expect, test, vi } from "vitest";

import { cleanup, fireEvent, render, screen, waitFor, within } from "../../lib/testing";
import Moderation from "./Moderation";
import ModerationDetail from "./ModerationDetail";

const reload = vi.fn();

vi.mock("@inertiajs/react", () => ({
	Head: () => null,
	Link: ({ href, children }: { href: string; children: React.ReactNode }) => (
		<a href={href}>{children}</a>
	),
	router: { reload: () => reload(), get: vi.fn() },
}));
vi.mock("../../lib/useT", () => ({
	useT: () => (key: string, p?: Record<string, string | number>) =>
		p ? `${key} ${JSON.stringify(p)}` : key,
}));

const fetchMock = vi.fn();

beforeEach(() => {
	fetchMock.mockResolvedValue({
		ok: true,
		status: 200,
		statusText: "OK",
		json: async () => ({}),
	});
	vi.stubGlobal("fetch", fetchMock);
});
afterEach(() => {
	cleanup();
	fetchMock.mockReset();
	reload.mockReset();
	vi.unstubAllGlobals();
});

const row = (over = {}) => ({
	id: "S1",
	kind: "edit",
	targetType: "item",
	targetID: "T1",
	domain: "movie",
	locale: "en",
	status: "submitted",
	fields: ["title"],
	authorID: 7,
	assignedTo: 0,
	authorTier: "trusted",
	authorAccuracy: 0.9,
	version: 1,
	age: "amber",
	competing: false,
	stale: true,
	unreachable: false,
	recApprove: 0,
	recReject: 1,
	...over,
});

const queue = (rows: ReturnType<typeof row>[]) => ({
	rows,
	filter: {
		status: "",
		kind: "",
		domain: "",
		age: "",
		field: "",
		competing: false,
		stale: false,
	},
	nextUrl: "",
	statuses: ["submitted", "in_review"],
	kinds: ["edit"],
	domains: ["movie"],
	ageBuckets: ["amber"],
});

test("the queue lists a suggestion with its badges and recommendation counts", () => {
	render(<Moderation queue={queue([row()])} canApply={false} rejectReasons={null} />);
	const r = screen.getByRole("row", { name: /item · movie · en/ });
	expect(within(r).getByText("amber")).toBeTruthy();
	expect(within(r).getByText("admin.suggestion.stale")).toBeTruthy();
	expect(within(r).getByText(/admin.moderation.recs/)).toBeTruthy();
	expect(screen.queryByText("admin.moderation.approve")).toBeNull();
});

test("claiming posts the live version to the claim endpoint", async () => {
	render(<Moderation queue={queue([row()])} canApply={false} rejectReasons={null} />);
	fireEvent.click(screen.getByRole("button", { name: "admin.moderation.claim" }));
	await waitFor(() => expect(fetchMock).toHaveBeenCalled());
	expect(fetchMock.mock.calls[0][0]).toBe("/admin/moderation/S1/claim");
	expect((fetchMock.mock.calls[0][1].body as URLSearchParams).get("version")).toBe("1");
});

test("a suggestion in review offers release, and admins get the bulk approve and reject", () => {
	render(
		<Moderation
			queue={queue([row({ status: "in_review" })])}
			canApply
			rejectReasons={["spam"]}
		/>,
	);
	expect(screen.getByRole("button", { name: "admin.moderation.release" })).toBeTruthy();
	expect(screen.getByRole("option", { name: "admin.moderation.approve" })).toBeTruthy();
	expect(screen.getByRole("option", { name: "spam" })).toBeTruthy();
});

const detail = (over = {}) => ({
	id: "S1",
	kind: "edit",
	targetType: "item",
	targetID: "T1",
	domain: "movie",
	locale: "en",
	status: "submitted",
	note: "",
	rejectReason: "",
	decisionNote: "",
	sourceURLs: null,
	version: 2,
	authorID: 7,
	age: "green",
	competing: false,
	stale: false,
	unreachable: false,
	diffs: [
		{
			field: "title.en",
			kind: "text",
			current: "Old",
			before: "Old",
			after: "<script>alert(1)</script> New",
			words: [
				{ op: "del", text: "Old" },
				{ op: "add", text: "<script>alert(1)</script>" },
			],
			add: null,
			remove: null,
		},
	],
	reviews: [
		{
			reviewerId: 5,
			recommendation: "reject",
			note: "Wrong year",
			version: 2,
			at: "2026-11-03T10:00:00Z",
		},
	],
	...over,
});

test("the diff and the recommendation render as text, never as markup", () => {
	const { container } = render(
		<ModerationDetail detail={detail()} canApply rejectReasons={["spam"]} />,
	);
	expect(screen.getByText("<script>alert(1)</script>")).toBeTruthy();
	expect(container.querySelector("script")).toBeNull();
	expect(screen.getByText(/admin.suggestion.recommends/)).toBeTruthy();
});

test("an admin can approve with the live version; a merge asks for the typed title", async () => {
	const { rerender } = render(
		<ModerationDetail detail={detail()} canApply rejectReasons={["spam"]} />,
	);
	expect(screen.queryByLabelText("admin.suggestion.mergeConfirm")).toBeNull();
	rerender(
		<ModerationDetail detail={detail({ kind: "merge" })} canApply rejectReasons={["spam"]} />,
	);
	expect(screen.getByLabelText("admin.suggestion.mergeConfirm")).toBeTruthy();
	fireEvent.change(screen.getByLabelText("admin.suggestion.mergeConfirm"), {
		target: { value: "Keeper" },
	});
	fireEvent.click(screen.getByRole("button", { name: "admin.suggestion.approveBtn" }));
	await waitFor(() => expect(fetchMock).toHaveBeenCalled());
	const body = fetchMock.mock.calls[0][1].body as URLSearchParams;
	expect(fetchMock.mock.calls[0][0]).toBe("/admin/moderation/S1/approve");
	expect(body.get("version")).toBe("2");
	expect(body.get("confirm_title")).toBe("Keeper");
});

test("a decided suggestion offers a clone and no decision forms", () => {
	render(
		<ModerationDetail detail={detail({ status: "rejected" })} canApply rejectReasons={null} />,
	);
	expect(screen.getByRole("button", { name: "admin.suggestion.clone" })).toBeTruthy();
	expect(screen.queryByRole("button", { name: "admin.suggestion.approveBtn" })).toBeNull();
});
