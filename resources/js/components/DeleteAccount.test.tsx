import { act, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, beforeEach, expect, test, vi } from "vitest";

import DeleteAccount from "./DeleteAccount";

const post = vi.fn();

vi.mock("@inertiajs/react", () => ({ router: { post: (...args: unknown[]) => post(...args) } }));
vi.mock("../lib/useT", () => ({
	useT: () => (key: string, params?: Record<string, number>) =>
		params ? `${key} ${params.seconds}` : key,
}));

beforeEach(() => vi.useFakeTimers({ shouldAdvanceTime: true }));
afterEach(() => {
	vi.useRealTimers();
	post.mockClear();
});

test("confirm stays disabled for 10 seconds, then posts the confirmation", async () => {
	render(<DeleteAccount />);
	await userEvent.click(screen.getByRole("button", { name: "account.delete.button" }));

	const wait = screen.getByRole("button", { name: "account.delete.wait 10" });
	expect(wait).toHaveProperty("disabled", true);

	await act(async () => {
		vi.advanceTimersByTime(10_000);
	});
	const confirm = screen.getByRole("button", { name: "account.delete.confirm" });
	expect(confirm).toHaveProperty("disabled", false);

	await userEvent.click(confirm);
	expect(post).toHaveBeenCalledWith("/profile/deletion", { confirm: true });
});
