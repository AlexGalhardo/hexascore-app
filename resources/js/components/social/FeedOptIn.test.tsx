import { cleanup, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, expect, test, vi } from "vitest";

import FeedOptIn from "./FeedOptIn";

const patch = vi.fn();

vi.mock("@inertiajs/react", () => ({
	router: { patch: (...args: unknown[]) => patch(...args) },
}));
vi.mock("../../lib/useLocalized", () => ({
	useLocalized: () => (path: string) => `/pt-BR${path}`,
}));
vi.mock("../../lib/useT", () => ({ useT: () => (key: string) => key }));

afterEach(() => {
	cleanup();
	patch.mockClear();
});

test("opting in saves at once on the localized route", async () => {
	render(<FeedOptIn on={false} />);

	const box = screen.getByRole("checkbox", { name: /social.optIn.label/ });
	expect((box as HTMLInputElement).checked).toBe(false);
	await userEvent.click(box);
	expect(patch.mock.calls[0][0]).toBe("/pt-BR/profile/feed");
	expect(patch.mock.calls[0][1]).toEqual({ feedOptIn: true });
});

test("opting out sends false", async () => {
	render(<FeedOptIn on />);

	await userEvent.click(screen.getByRole("checkbox", { name: /social.optIn.label/ }));
	expect(patch.mock.calls[0][1]).toEqual({ feedOptIn: false });
});
