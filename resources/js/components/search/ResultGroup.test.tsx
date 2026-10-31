import { cleanup, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { ReactNode } from "react";
import { afterEach, expect, test, vi } from "vitest";

import ResultGroup from "./ResultGroup";

vi.mock("@inertiajs/react", () => ({
	Link: ({
		children,
		href,
		onClick,
	}: {
		children: ReactNode;
		href: string;
		onClick?: () => void;
	}) => (
		<a
			href={href}
			onClick={(e) => {
				e.preventDefault();
				onClick?.();
			}}
		>
			{children}
		</a>
	),
	usePage: () => ({ props: { coversBase: "" } }),
}));

afterEach(cleanup);

test("click reports only the shown ref with a 1-based position in the group", async () => {
	const beacon = vi.fn();
	vi.stubGlobal("navigator", { sendBeacon: beacon });
	const hits = [
		{ id: "a", type: "game", title: "Zelda", path: "/games/zelda" },
		{ id: "b", type: "game", title: "Zelda II", path: "/games/zelda-ii" },
	];
	render(<ResultGroup hits={hits} searchId="s1" group="game" q="zel" offset={20} />);
	await userEvent.click(screen.getByRole("link", { name: "Zelda II" }));
	expect(beacon).toHaveBeenCalledTimes(1);
	const [url, blob] = beacon.mock.calls[0] as [string, Blob];
	expect(url).toBe("/search/events");
	expect(JSON.parse(await blob.text())).toEqual({
		searchId: "s1",
		refId: "b",
		position: 22,
		group: "game",
	});
	vi.unstubAllGlobals();
});
