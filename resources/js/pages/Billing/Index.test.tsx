import { cleanup, render, screen } from "@testing-library/react";
import type { ReactNode } from "react";
import { afterEach, expect, test, vi } from "vitest";

import Billing from "./Index";

vi.mock("../../lib/useT", () => ({
	useT: () => (key: string, args?: Record<string, unknown>) =>
		args ? `${key} ${JSON.stringify(args)}` : key,
}));
vi.mock("@inertiajs/react", () => ({
	Head: () => null,
	Link: ({ children }: { children: ReactNode }) => <button type="button">{children}</button>,
	usePage: () => ({ props: { locale: "en" } }),
}));

afterEach(cleanup);

const sub = (state: string, cancelAtPeriodEnd = false) => ({
	plan: "Hobby",
	status: state,
	state,
	valid: state === "active",
	currentPeriodEnd: "2026-12-02",
	cancelAtPeriodEnd,
});

const render_ = (s: ReturnType<typeof sub>) =>
	render(
		<Billing plans={[]} subscription={s} subscribed={false} checkout={false} currency="usd" />,
	);

test("an active subscription shows its renewal date", () => {
	render_(sub("active"));
	expect(screen.getByText(/billing\.checkout\.renews/)).toBeTruthy();
});

test("a subscription set to end shows the end date", () => {
	render_(sub("active", true));
	expect(screen.getByText(/billing\.checkout\.ends/)).toBeTruthy();
});

test("a canceled subscription shows no renewal or end date", () => {
	render_(sub("canceled"));
	expect(screen.queryByText(/billing\.checkout\.(renews|ends)/)).toBeNull();
});
