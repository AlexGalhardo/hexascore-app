import { cleanup, render, screen } from "@testing-library/react";
import { afterEach, expect, test, vi } from "vitest";

import RatingPanel from "./RatingPanel";

vi.mock("@inertiajs/react", () => ({
	Link: ({ children, href }: { children: React.ReactNode; href: string }) => (
		<a href={href}>{children}</a>
	),
	router: { put: () => undefined, delete: () => undefined },
	usePage: () => ({ props: { locale: "en" } }),
}));
vi.mock("../../lib/useT", () => ({ useT: () => (key: string) => key }));

afterEach(cleanup);

const panel = (canWrite: boolean) =>
	render(<RatingPanel itemId="i1" type="movie" mine={null} canWrite={canWrite} />);

test("the card is headed 'Your rating' only; the old 'Rate <item>' heading is gone", () => {
	panel(true);

	const heading = screen.getByRole("heading", { level: 2 });
	expect(heading.textContent).toBe("item.rate");
	expect(screen.getByRole("region", { name: "item.rate" })).toBeTruthy();
	expect(screen.queryByText("item.panel.title")).toBeNull();
});

test("the help text is visually hidden and describes every star group", () => {
	panel(true);

	const groups = screen.getAllByRole("radiogroup");
	expect(groups).toHaveLength(6);
	for (const g of groups) {
		const id = g.getAttribute("aria-describedby") ?? "";
		const help = document.getElementById(id);
		expect(help?.textContent).toBe("item.panel.help");
		expect(help?.className).toContain("sr-only");
	}
});

test("guests get the guest help text as the description", () => {
	panel(false);

	const id = screen.getAllByRole("radiogroup")[0].getAttribute("aria-describedby") ?? "";
	expect(document.getElementById(id)?.textContent).toBe("item.panel.helpGuest");
});
