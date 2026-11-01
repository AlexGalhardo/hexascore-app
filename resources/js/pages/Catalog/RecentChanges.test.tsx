import { cleanup, render, screen } from "@testing-library/react";
import type { ReactNode } from "react";
import { afterEach, expect, test, vi } from "vitest";

import RecentChanges from "./RecentChanges";

vi.mock("../../lib/useT", () => ({
	useT: () => (key: string) => key,
}));
vi.mock("../../components/Seo", () => ({ default: () => null }));
vi.mock("@inertiajs/react", () => ({
	Link: ({ href, children }: { href: string; children: ReactNode }) => (
		<a href={href}>{children}</a>
	),
	usePage: () => ({ url: "/movies/old/changes?page=2", props: { locale: "en" } }),
}));

afterEach(cleanup);

const at = "2026-10-31T10:00:00Z";

test("renders actors as nickname, team or deleted, values as text, and paging", () => {
	render(
		<RecentChanges
			target={{ type: "item", id: "i1", title: "Old" }}
			page={2}
			nextPage={3}
			entries={[
				{
					id: 3,
					occurredAt: at,
					source: "community",
					field: "title",
					before: '"a"',
					after: '"<b>x</b>"',
					actor: "user",
					nickname: "Ana",
				},
				{
					id: 2,
					occurredAt: at,
					source: "admin",
					field: "release_date",
					before: "null",
					after: '{"value":"2001-02-03"}',
					actor: "team",
				},
				{
					id: 1,
					occurredAt: at,
					source: "merge",
					field: "*merge*",
					before: "",
					after: "",
					actor: "deleted",
				},
			]}
		/>,
	);

	expect(screen.getByRole("link", { name: "Ana" }).getAttribute("href")).toBe("/u/ana");
	expect(screen.getByText("community.recent.actors.team")).toBeTruthy();
	expect(screen.getByText("community.recent.actors.deleted")).toBeTruthy();
	expect(screen.queryByText("*merge*")).toBeNull();
	expect(screen.getByText("<b>x</b>")).toBeTruthy();
	expect(screen.getByText('{"value":"2001-02-03"}')).toBeTruthy();
	expect(screen.getByText("—")).toBeTruthy();
	expect(screen.getByRole("link", { name: "item.previous" }).getAttribute("href")).toBe(
		"/movies/old/changes?page=1",
	);
	expect(screen.getByRole("link", { name: "item.next" }).getAttribute("href")).toBe(
		"/movies/old/changes?page=3",
	);
	expect(
		screen.getByRole("link", { name: "community.recent.overview" }).getAttribute("href"),
	).toBe("/movies/old");
});

test("shows a column header row and the empty state", () => {
	const { unmount } = render(
		<RecentChanges
			target={{ type: "item", id: "i1", title: "Old" }}
			page={1}
			nextPage={null}
			entries={[]}
		/>,
	);
	expect(screen.getByText("community.recent.empty")).toBeTruthy();
	expect(screen.queryByRole("table")).toBeNull();
	unmount();
	render(
		<RecentChanges
			target={{ type: "item", id: "i1", title: "Old" }}
			page={1}
			nextPage={null}
			entries={[
				{
					id: 1,
					occurredAt: at,
					source: "community",
					field: "x_col",
					before: "1",
					after: "2",
					actor: "system",
				},
			]}
		/>,
	);
	expect(
		screen.getByRole("columnheader", { name: "community.recent.columns.field" }),
	).toBeTruthy();
	expect(screen.getByRole("cell", { name: "x_col" })).toBeTruthy();
	expect(screen.queryByRole("link", { name: "item.previous" })).toBeNull();
});
