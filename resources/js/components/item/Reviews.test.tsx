import { cleanup, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, expect, test, vi } from "vitest";

import Reviews, { type ReviewEntry } from "./Reviews";

const put = vi.fn();
const del = vi.fn();

vi.mock("@inertiajs/react", () => ({
	router: {
		put: (...args: unknown[]) => put(...args),
		delete: (...args: unknown[]) => del(...args),
	},
}));
vi.mock("../../lib/useLocalized", () => ({ useLocalized: () => (path: string) => path }));
vi.mock("../../lib/useT", () => ({ useT: () => (key: string) => key }));

afterEach(() => {
	cleanup();
	put.mockClear();
	del.mockClear();
});

const entry = (over: Partial<ReviewEntry>): ReviewEntry => ({
	id: "r1",
	author: { nickname: "Critic", initial: "C" },
	stars: 8,
	body: [{ t: "p", c: [{ t: "text", text: "Plot twist inside" }] }],
	spoiler: false,
	helpful: 0,
	createdAt: "2026-11-01T10:00:00Z",
	updatedAt: "2026-11-01T10:00:00Z",
	...over,
});

test("a spoiler stays hidden until revealed", async () => {
	render(
		<Reviews
			itemId="i1"
			reviews={{ data: [entry({ spoiler: true })], mine: null }}
			canWrite={false}
			rated={false}
		/>,
	);

	expect(screen.queryByText("Plot twist inside")).toBeNull();
	await userEvent.click(screen.getByRole("button", { name: "reviews.reveal" }));
	expect(screen.getByText("Plot twist inside")).toBeTruthy();
});

test("a rated member writes a review with the spoiler flag", async () => {
	render(<Reviews itemId="i1" reviews={{ data: [], mine: null }} canWrite rated />);

	expect(screen.getByText("reviews.empty")).toBeTruthy();
	await userEvent.type(screen.getByLabelText("reviews.body"), "Loved it");
	await userEvent.click(screen.getByLabelText("reviews.spoiler"));
	await userEvent.click(screen.getByRole("button", { name: "reviews.save" }));
	expect(put.mock.calls[0][0]).toBe("/items/i1/review");
	expect(put.mock.calls[0][1]).toEqual({ body: "Loved it", spoiler: true });
});

test("the author edits and deletes their review; an unrated member is asked to rate first", async () => {
	const mine = entry({ source: "**old**" });
	const { unmount } = render(
		<Reviews itemId="i1" reviews={{ data: [mine], mine }} canWrite rated />,
	);

	expect((screen.getByLabelText("reviews.body") as HTMLTextAreaElement).value).toBe("**old**");
	await userEvent.click(screen.getByRole("button", { name: "reviews.delete" }));
	expect(del.mock.calls[0][0]).toBe("/items/i1/review");
	unmount();

	render(<Reviews itemId="i1" reviews={{ data: [], mine: null }} canWrite rated={false} />);
	expect(screen.getByText("reviews.rateFirst")).toBeTruthy();
	expect(screen.queryByLabelText("reviews.body")).toBeNull();
});
