import { afterEach, expect, test, vi } from "vitest";

import type { OverviewData } from "../../components/admin/overview";
import { cleanup, render, screen, within } from "../../lib/testing";
import Dashboard from "./Dashboard";

vi.mock("@inertiajs/react", () => ({
	Head: () => null,
	usePage: () => ({ url: "/admin/overview", props: { locale: "en" } }),
	Link: ({ href, children }: { href: string; children: React.ReactNode }) => (
		<a href={href}>{children}</a>
	),
}));
vi.mock("../../lib/useT", () => ({
	useT: () => (key: string, p?: Record<string, string>) =>
		p ? `${key} ${JSON.stringify(p)}` : key,
}));

afterEach(cleanup);

const overview = (over: Partial<OverviewData> = {}): OverviewData => ({
	generatedAt: new Date().toISOString(),
	counts: {
		items: 1234,
		users: 56,
		openReports: 3,
		runningRuns: 1,
		failedRuns24h: 0,
		parkedFailures: 2,
	},
	worker: { status: "ok", ageSeconds: 42 },
	runs: [
		{
			id: 7,
			job: "fetch.wikidata",
			source: "wikidata",
			mode: "bootstrap",
			status: "succeeded",
			seen: 100,
			changed: 40,
			failed: 0,
			startedAt: "2026-11-03T20:00:00Z",
			durationSeconds: 83,
		},
	],
	...over,
});

test("shows the headline numbers and the recent runs from the first snapshot", () => {
	render(<Dashboard overview={overview()} />);
	expect(screen.getByText("1,234")).toBeTruthy();
	expect(screen.getByText("56")).toBeTruthy();
	const row = screen.getByRole("row", { name: /fetch.wikidata/ });
	expect(within(row).getByText("admin.dashboard.runs.statuses.succeeded")).toBeTruthy();
	expect(within(row).getByText("1m 23s")).toBeTruthy();
});

test("a healthy worker shows its heartbeat age", () => {
	render(<Dashboard overview={overview()} />);
	expect(screen.getByText("admin.dashboard.worker.ok")).toBeTruthy();
	expect(screen.getByText(/admin.dashboard.worker.age.*42s/)).toBeTruthy();
});

test("a stale or never-started worker says so, with the way to investigate", () => {
	render(<Dashboard overview={overview({ worker: { status: "stale", ageSeconds: 900 } })} />);
	expect(screen.getByText("admin.dashboard.worker.stale")).toBeTruthy();
	expect(screen.getByText("admin.dashboard.worker.staleHint")).toBeTruthy();
	cleanup();
	render(<Dashboard overview={overview({ worker: { status: "never", ageSeconds: null } })} />);
	expect(screen.getByText("admin.dashboard.worker.neverHint")).toBeTruthy();
});

test("no runs shows the empty message instead of an empty table", () => {
	render(<Dashboard overview={overview({ runs: [] })} />);
	expect(screen.getByText("admin.dashboard.runs.empty")).toBeTruthy();
	expect(screen.queryByRole("table")).toBeNull();
});
