import userEvent from "@testing-library/user-event";
import type { ReactNode } from "react";
import { expect, test, vi } from "vitest";
import { render, screen } from "../../lib/testing";

import Edit from "./Edit";

const put = vi.hoisted(() => vi.fn());
vi.mock("../../lib/useT", () => ({ useT: () => (key: string) => key }));
vi.mock("../../components/Seo", () => ({ default: () => null }));
vi.mock("../../components/search/SuggestBox", () => ({
	default: ({ onPick, type }: { onPick: (s: { id: string }) => void; type: string }) => (
		<button type="button" onClick={() => onPick({ id: "new-item" })}>
			suggest-{type}
		</button>
	),
}));
vi.mock("@inertiajs/react", () => ({
	Link: ({ href, children }: { href: string; children: ReactNode }) => (
		<a href={href}>{children}</a>
	),
	router: { delete: vi.fn(), put },
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
			list={{
				id: "l1",
				type: "game",
				title: "L",
				description: null,
				isPublic: true,
				hidden: false,
			}}
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

test("picking a suggestion puts the item into the list", async () => {
	render(
		<Edit
			list={{
				id: "l1",
				type: "game",
				title: "L",
				description: null,
				isPublic: true,
				hidden: false,
			}}
			items={[]}
			version={1}
		/>,
	);
	await userEvent
		.setup()
		.click(screen.getAllByRole("button", { name: "suggest-game" }).at(-1) as HTMLElement);
	expect(put).toHaveBeenCalledWith("/lists/l1/items/new-item", {}, { preserveScroll: true });
});
