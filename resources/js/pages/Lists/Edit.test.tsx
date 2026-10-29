import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { ReactNode } from "react";
import { expect, test, vi } from "vitest";

import Edit from "./Edit";

vi.mock("../../lib/useT", () => ({ useT: () => (key: string) => key }));
vi.mock("../../components/Seo", () => ({ default: () => null }));
vi.mock("@inertiajs/react", () => ({
	Link: ({ href, children }: { href: string; children: ReactNode }) => (
		<a href={href}>{children}</a>
	),
	router: { delete: vi.fn(), put: vi.fn() },
	usePage: () => ({ props: { locale: "en" } }),
	useForm: (data: Record<string, unknown>) => ({
		data,
		errors: {},
		processing: false,
		setData: vi.fn(),
		patch: vi.fn(),
	}),
}));

const entry = (id: string, position: number) => ({
	position,
	note: null,
	item: { id, type: "game" as const, slug: id, title: id, year: null, coverThumbKey: null },
});

test("keyboard reorder keeps focus on the moved row", async () => {
	const user = userEvent.setup();
	render(
		<Edit
			list={{ id: "l1", title: "L", description: null, isPublic: true, hidden: false }}
			items={[entry("a", 1), entry("b", 2), entry("c", 3)]}
			version={1}
		/>,
	);

	screen.getByRole("button", { name: "userLists.moveDown: a" }).focus();
	await user.keyboard("{Enter}");
	expect(document.activeElement?.id).toBe("move-down-a");

	await user.keyboard("{Enter}");
	// "a" is now last, so its move-down button is disabled and focus moves to move-up.
	expect(document.activeElement?.id).toBe("move-up-a");
	expect(screen.getAllByRole("listitem").map((li) => li.textContent?.slice(1, 2))).toEqual([
		"b",
		"c",
		"a",
	]);
});
