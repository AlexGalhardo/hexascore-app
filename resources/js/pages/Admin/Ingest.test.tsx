import { afterEach, beforeEach, expect, test, vi } from "vitest";

import type { IngestStatus, RunRow } from "../../components/admin/types";
import { cleanup, fireEvent, render, screen, waitFor, within } from "../../lib/testing";
import Ingest from "./Ingest";

const reload = vi.fn();

vi.mock("@inertiajs/react", () => ({
	Head: () => null,
	Link: ({
		href,
		children,
		className,
	}: {
		href: string;
		children: React.ReactNode;
		className?: string;
	}) => (
		<a href={href} className={className}>
			{children}
		</a>
	),
	router: { reload: () => reload() },
}));
vi.mock("../../lib/useT", () => ({
	useT: () => (key: string, p?: Record<string, string | number>) =>
		p ? `${key} ${JSON.stringify(p)}` : key,
}));

const fetchMock = vi.fn();

beforeEach(() => {
	fetchMock.mockResolvedValue({
		ok: true,
		status: 200,
		statusText: "OK",
		json: async () => ({}),
	});
	vi.stubGlobal("fetch", fetchMock);
});
afterEach(() => {
	cleanup();
	fetchMock.mockReset();
	reload.mockReset();
	vi.unstubAllGlobals();
});

const run = (over: Partial<RunRow> = {}): RunRow => ({
	id: 7,
	job: "fetch.wikidata",
	source: "wikidata",
	status: "running",
	error: "",
	started: "2026-11-03 20:00",
	finished: "",
	seen: 100,
	changed: 40,
	failed: 1,
	rollbackable: false,
	...over,
});

const ingest = (over: Partial<IngestStatus> = {}): IngestStatus => ({
	failures: null,
	asserts: null,
	oldestTMDB: "",
	coverage: null,
	stages: [
		{ job: "fetch.wikidata", source: "wikidata" },
		{ job: "normalize", source: "" },
	],
	modes: ["bootstrap", "manual"],
	canRun: true,
	canRollback: true,
	...over,
});

test("lists the runs and the empty states of the other sections", () => {
	render(<Ingest runs={[run()]} ingest={ingest()} />);
	const row = screen.getByRole("row", { name: /fetch.wikidata/ });
	expect(within(row).getByText("admin.dashboard.runs.statuses.running")).toBeTruthy();
	expect(screen.getByText("admin.ingest.noAsserts")).toBeTruthy();
	expect(screen.getByText("admin.ingest.noFailures")).toBeTruthy();
	expect(screen.getByText("admin.ingest.noTmdb")).toBeTruthy();
	expect(screen.getByText("admin.ingest.noCoverage")).toBeTruthy();
});

test("a run in progress can be cancelled and the page reloads", async () => {
	render(<Ingest runs={[run()]} ingest={ingest()} />);
	fireEvent.click(screen.getByRole("button", { name: "admin.ingest.cancelRun" }));
	await waitFor(() => expect(reload).toHaveBeenCalled());
	expect(fetchMock.mock.calls[0][0]).toBe("/admin/ingest/runs/7/cancel");
	expect(fetchMock.mock.calls[0][1].method).toBe("POST");
});

test("without the run ability there is no start form and no cancel button", () => {
	render(<Ingest runs={[run()]} ingest={ingest({ canRun: false })} />);
	expect(screen.queryByText("admin.ingest.startRun")).toBeNull();
	expect(screen.queryByRole("button", { name: "admin.ingest.cancelRun" })).toBeNull();
});

test("starting a run posts the form-encoded fields, the checkbox only when ticked", async () => {
	render(<Ingest runs={[]} ingest={ingest()} />);
	expect(screen.getByText("admin.ingest.noRuns")).toBeTruthy();
	fireEvent.change(screen.getByLabelText("admin.ingest.limit"), { target: { value: "100" } });
	fireEvent.click(screen.getByLabelText("admin.ingest.allowDrop"));
	fireEvent.click(screen.getByRole("button", { name: "admin.ingest.start" }));
	await waitFor(() => expect(fetchMock).toHaveBeenCalled());
	const [url, init] = fetchMock.mock.calls[0];
	expect(url).toBe("/admin/ingest/runs");
	const body = init.body as URLSearchParams;
	expect(body.get("job")).toBe("fetch.wikidata");
	expect(body.get("mode")).toBe("bootstrap");
	expect(body.get("limit")).toBe("100");
	expect(body.get("allow_count_drop")).toBe("on");
});

test("a refused action shows the server message and does not reload", async () => {
	fetchMock.mockResolvedValue({
		ok: false,
		status: 409,
		statusText: "Conflict",
		json: async () => ({ detail: "A run of this job is already active." }),
	});
	render(<Ingest runs={[]} ingest={ingest()} />);
	fireEvent.click(screen.getByRole("button", { name: "admin.ingest.start" }));
	expect(await screen.findByText("A run of this job is already active.")).toBeTruthy();
	expect(reload).not.toHaveBeenCalled();
});
