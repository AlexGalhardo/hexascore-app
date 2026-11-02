import { cleanup, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, expect, test, vi } from "vitest";

import TimeToBeat from "./TimeToBeat";

const put = vi.fn();
const del = vi.fn();

vi.mock("@inertiajs/react", () => ({
	router: {
		put: (...args: unknown[]) => {
			put(...args);
			(args[2] as { onFinish?: () => void }).onFinish?.();
		},
		delete: (...args: unknown[]) => del(...args),
	},
	Link: ({ href, children }: { href: string; children: React.ReactNode }) => (
		<a href={href}>{children}</a>
	),
}));
vi.mock("../../lib/useLocalized", () => ({ useLocalized: () => (path: string) => path }));
vi.mock("../../lib/useT", () => ({
	useT: () => (key: string, p?: Record<string, number>) =>
		p ? `${key} ${p.hours ?? p.count}` : key,
}));

afterEach(() => {
	cleanup();
	put.mockClear();
	del.mockClear();
});

const medians = { mainHours: 12, extrasHours: null, completionistHours: 40, reports: 5 };

test("shows the median per style with the report count, skipping empty styles", () => {
	render(<TimeToBeat itemId="i1" medians={medians} mine={null} canEdit={false} />);
	expect(screen.getByText("item.timeToBeat.hours 12")).toBeTruthy();
	expect(screen.getByText("item.timeToBeat.hours 40")).toBeTruthy();
	expect(screen.queryByText("item.timeToBeat.styles.extras")).toBeNull();
	expect(screen.getByText(/item.timeToBeat.reports 5/)).toBeTruthy();
});

test("signed out: no form, a log in link", () => {
	render(<TimeToBeat itemId="i1" medians={medians} mine={null} canEdit={false} />);
	expect(screen.queryByRole("button", { name: "item.timeToBeat.save" })).toBeNull();
	expect(
		screen.getByRole("link", { name: "item.timeToBeat.loginToReport" }).getAttribute("href"),
	).toBe("/login");
});

test("no reports: empty message; signed in can still submit, and clear their own entry", async () => {
	const empty = { mainHours: null, extrasHours: null, completionistHours: null, reports: 0 };
	render(
		<TimeToBeat
			itemId="i1"
			medians={empty}
			mine={{ mainHours: 10, extrasHours: null, completionistHours: null }}
			canEdit
		/>,
	);
	expect(screen.getByText("item.timeToBeat.none")).toBeTruthy();

	await userEvent.click(screen.getByRole("button", { name: "item.timeToBeat.save" }));
	expect(put).toHaveBeenCalledWith(
		"/items/i1/time-to-beat",
		{ mainHours: 10, extrasHours: null, completionistHours: null },
		expect.anything(),
	);

	await userEvent.click(screen.getByRole("button", { name: "item.timeToBeat.delete" }));
	expect(del).toHaveBeenCalledWith("/items/i1/time-to-beat", expect.anything());
});
