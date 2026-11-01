import { cleanup, render, screen } from "@testing-library/react";
import type { ReactNode } from "react";
import { afterEach, expect, test, vi } from "vitest";

import ActivityCard, { type ActivityEntry } from "./ActivityCard";

vi.mock("../../lib/useT", () => ({
	useT: () => (key: string, p?: Record<string, unknown>) =>
		p ? `${key}:${Object.values(p).join(",")}` : key,
}));
vi.mock("@inertiajs/react", () => ({
	Link: ({ href, children }: { href: string; children: ReactNode }) => (
		<a href={href}>{children}</a>
	),
	usePage: () => ({ props: { locale: "en", coversBase: "/covers" } }),
}));

afterEach(cleanup);

const base = {
	id: "e1",
	actor: { nickname: "OpenA", initial: "O" },
	payload: {},
	at: "2026-11-01T12:00:00Z",
};

test("a rating names the actor, the item and the stars", () => {
	const entry: ActivityEntry = {
		...base,
		kind: "rating_saved",
		item: { id: "i1", type: "game", slug: "zelda", title: "Zelda", cover: null, stars: 9 },
	};
	render(<ActivityCard entry={entry} />);

	expect(screen.getByText("social.feed.rating_saved:OpenA")).toBeTruthy();
	expect(screen.getByRole("link", { name: "OpenA" }).getAttribute("href")).toBe("/u/opena");
	expect(screen.getByText("Zelda")).toBeTruthy();
	// The actor's score, not ShelfCard's "Your score".
	expect(screen.getByText("social.feed.score:9.0")).toBeTruthy();
	expect(screen.queryByText(/library.yourScore/)).toBeNull();
});

test("a status change prints the status label; a list links to the list", () => {
	render(
		<ActivityCard
			entry={{
				...base,
				kind: "status_changed",
				payload: { status: "completed" },
				item: { id: "i2", type: "game", slug: "ico", title: "Ico", cover: null },
			}}
		/>,
	);
	expect(
		screen.getByText("social.feed.status_changed:OpenA,social.status.game.completed"),
	).toBeTruthy();
	cleanup();

	render(
		<ActivityCard
			entry={{ ...base, kind: "list_created", list: { id: "l1", title: "Shared" } }}
		/>,
	);
	expect(screen.getByRole("link", { name: "Shared" }).getAttribute("href")).toBe(
		"/u/opena/lists/l1",
	);
});
