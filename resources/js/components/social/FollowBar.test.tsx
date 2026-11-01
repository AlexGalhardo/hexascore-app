import { cleanup, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, expect, test, vi } from "vitest";

import FollowBar from "./FollowBar";

const put = vi.fn();
const del = vi.fn();

vi.mock("@inertiajs/react", () => ({
	router: {
		put: (...args: unknown[]) => put(...args),
		delete: (...args: unknown[]) => del(...args),
	},
}));
vi.mock("../../lib/useLocalized", () => ({ useLocalized: () => (path: string) => path }));
vi.mock("../../lib/useT", () => ({
	useT: () => (key: string, params?: Record<string, number>) =>
		params ? `${key} ${params.count}` : key,
}));

afterEach(() => {
	cleanup();
	put.mockClear();
	del.mockClear();
});

test("shows the counts and follows a profile", async () => {
	render(
		<FollowBar nickname="StarB" social={{ followers: 3, following: 1, state: "" }} canFollow />,
	);

	expect(screen.getByText("social.follow.followers 3")).toBeTruthy();
	expect(screen.getByText("social.follow.following 1")).toBeTruthy();
	await userEvent.click(screen.getByRole("button", { name: "social.follow.follow" }));
	expect(put.mock.calls[0][0]).toBe("/u/starb/follow");
});

test("a pending request can be cancelled and hides nothing it does not have", async () => {
	render(<FollowBar nickname="LockedP" social={{ state: "pending" }} canFollow />);

	expect(screen.queryByText(/social.follow.followers/)).toBeNull();
	await userEvent.click(screen.getByRole("button", { name: "social.follow.requested" }));
	expect(del.mock.calls[0][0]).toBe("/u/lockedp/follow");
});

test("an accepted follow unfollows; anonymous viewers get no button", async () => {
	const { rerender } = render(
		<FollowBar
			nickname="StarB"
			social={{ followers: 1, following: 0, state: "accepted" }}
			canFollow
		/>,
	);
	await userEvent.click(screen.getByRole("button", { name: "social.follow.unfollow" }));
	expect(del).toHaveBeenCalledTimes(1);

	rerender(
		<FollowBar nickname="StarB" social={{ followers: 1, following: 0 }} canFollow={false} />,
	);
	expect(screen.queryByRole("button")).toBeNull();
});
