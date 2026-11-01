import { cleanup, render, screen, within } from "@testing-library/react";
import type { ReactNode } from "react";
import { afterEach, expect, test, vi } from "vitest";

import Index from "./Index";

vi.mock("../../lib/useT", () => ({ useT: () => (key: string) => key }));
vi.mock("../../components/Seo", () => ({ default: () => null }));
vi.mock("@inertiajs/react", () => ({
	Link: ({ children, ...rest }: { href: string; children: ReactNode }) => (
		<a {...rest}>{children}</a>
	),
	usePage: () => ({ props: { locale: "en", auth: { user: null } } }),
}));

afterEach(cleanup);

const plan = (id: "free" | "pro" | "business", cents: number) => ({
	id,
	amounts: { usd: cents },
	quota: 100,
	lifetime: id === "free",
	maxPageSize: 10,
	rps: 1,
});

test("prices in the chosen currency, marks pro as most chosen and the current currency", () => {
	render(
		<Index
			plans={[plan("free", 0), plan("pro", 600), plan("business", 2950)]}
			currency="usd"
			currencies={["usd", "brl"]}
			coverage={{ measured: false, docsPath: "/docs/coverage" }}
		/>,
	);
	const pro = screen.getByRole("listitem", { name: /billing\.plans\.pro/ });
	expect(within(pro).getByText("billing.pricing.popular")).toBeTruthy();
	expect(within(pro).getByText("$6")).toBeTruthy();
	expect(screen.getByText("$29.50")).toBeTruthy();
	const free = screen.getByRole("listitem", { name: /billing\.plans\.free/ });
	expect(within(free).getByText("billing.pricing.lifetime")).toBeTruthy();
	expect(within(free).queryByText("billing.pricing.popular")).toBeNull();
	expect(screen.getByRole("link", { name: "USD" }).getAttribute("aria-current")).toBe("true");
});
