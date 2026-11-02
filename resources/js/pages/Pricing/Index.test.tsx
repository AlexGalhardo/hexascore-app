import { cleanup, render, screen, within } from "@testing-library/react";
import type { ReactNode } from "react";
import { afterEach, expect, test, vi } from "vitest";

import Index from "./Index";

let user: { id: number } | null = null;

vi.mock("../../lib/useT", () => ({ useT: () => (key: string) => key }));
vi.mock("../../components/Seo", () => ({ default: () => null }));
vi.mock("@inertiajs/react", () => ({
	Link: ({
		children,
		href,
		method,
		data,
		className,
		"aria-current": current,
	}: {
		href: string;
		method?: string;
		data?: object;
		className?: string;
		"aria-current"?: "true";
		children: ReactNode;
	}) => (
		<a
			href={href}
			data-method={method}
			data-body={data ? JSON.stringify(data) : undefined}
			className={className}
			aria-current={current}
		>
			{children}
		</a>
	),
	usePage: () => ({ props: { locale: "en", auth: { user } } }),
}));

afterEach(() => {
	cleanup();
	user = null;
});

const plan = (id: "free" | "hobby" | "pro", cents: number) => ({
	id,
	amounts: { usd: cents, brl: cents * 4, eur: cents },
	quota: 100,
	lifetime: id === "free",
	maxPageSize: 10,
	rps: 1,
});

const plans = [plan("free", 0), plan("hobby", 500), plan("pro", 1000)];
const coverage = { docsPath: "/docs/coverage" };

test("prices each plan in the chosen currency and marks the current currency", () => {
	render(<Index plans={plans} currency="usd" currencies={["usd", "brl"]} coverage={coverage} />);
	const hobby = screen.getByRole("listitem", { name: /billing\.plans\.hobby/ });
	expect(within(hobby).getByText("$5")).toBeTruthy();
	expect(
		within(screen.getByRole("listitem", { name: /billing\.plans\.pro/ })).getByText("$10"),
	).toBeTruthy();
	const free = screen.getByRole("listitem", { name: /billing\.plans\.free/ });
	expect(within(free).getByText("billing.pricing.lifetime")).toBeTruthy();
	expect(screen.getByRole("link", { name: "USD" }).getAttribute("aria-current")).toBe("true");
});

test("signed in, each subscribe button posts straight to checkout with the selected currency", () => {
	user = { id: 1 };
	render(<Index plans={plans} currency="brl" currencies={["usd", "brl"]} coverage={coverage} />);
	for (const id of ["hobby", "pro"]) {
		const item = screen.getByRole("listitem", { name: new RegExp(`billing.plans.${id}`) });
		const cta = within(item).getByRole("link", { name: "billing.pricing.ctaPaid" });
		expect(cta.getAttribute("href")).toBe(`/billing/checkout/${id}`);
		expect(cta.getAttribute("data-method")).toBe("post");
		expect(cta.getAttribute("data-body")).toBe('{"currency":"brl"}');
	}
});

test("signed out, subscribing goes to signup and the page shows only the coverage link", () => {
	render(<Index plans={plans} currency="usd" currencies={["usd"]} coverage={coverage} />);
	const hobby = screen.getByRole("listitem", { name: /billing\.plans\.hobby/ });
	const cta = within(hobby).getByRole("link", { name: "billing.pricing.ctaSignup" });
	expect(cta.getAttribute("href")).toBe("/signup");
	expect(cta.getAttribute("data-method")).toBe("get");
	expect(screen.queryByText("billing.pricing.subtitle")).toBeNull();
	expect(screen.queryByText("billing.pricing.coverageBody")).toBeNull();
	expect(screen.queryByRole("complementary")).toBeNull();
	expect(
		screen.getByRole("link", { name: /billing\.pricing\.coverageLink/ }).getAttribute("href"),
	).toBe("/docs/coverage");
});
