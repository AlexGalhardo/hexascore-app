import { afterEach, beforeEach, expect, test, vi } from "vitest";

import type { ResColumn } from "../../components/admin/resources";
import { cleanup, fireEvent, render, screen, waitFor, within } from "../../lib/testing";
import Flags from "./Flags";
import ResourceList from "./ResourceList";
import ResourceShow from "./ResourceShow";

const get = vi.fn();
const reload = vi.fn();

vi.mock("@inertiajs/react", () => ({
	Head: () => null,
	Link: ({ href, children }: { href: string; children: React.ReactNode }) => (
		<a href={href}>{children}</a>
	),
	usePage: () => ({ url: "/admin/r/users", props: { locale: "en" } }),
	router: { get: (...a: unknown[]) => get(...a), reload: () => reload() },
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
	get.mockReset();
	reload.mockReset();
	fetchMock.mockReset();
	vi.unstubAllGlobals();
});

const columns: ResColumn[] = [
	{ key: "id", kind: "number", sortable: true },
	{ key: "email", kind: "text", sortable: true },
	{ key: "role", kind: "badge", sortable: false },
	{ key: "banned_at", kind: "time", sortable: false },
];

const base = {
	resource: { slug: "users", navKey: "users", hasDetail: true, searchable: true },
	columns,
	rows: [
		{ id: "7", cells: [7, "ada@example.com", "admin", null] },
		{ id: "8", cells: [8, "bob@example.com", "user", "2026-11-01T10:00:00Z"] },
	],
	filters: [{ key: "role", kind: "enum" as const, options: ["user", "admin"], value: "" }],
	page: 1,
	hasNext: true,
	q: "",
	sort: "",
	dir: "",
};

test("the list draws rows by kind, links the first cell to the record and offers the next page", () => {
	render(<ResourceList {...base} />);
	const row = screen.getByRole("row", { name: /ada@example.com/ });
	expect(within(row).getByText("admin")).toBeTruthy();
	expect(within(row).getAllByText("–").length).toBeGreaterThan(0);
	expect(within(row).getByRole("link", { name: "7" }).getAttribute("href")).toBe(
		"/admin/r/users/7",
	);
	expect(screen.getByRole("link", { name: "admin.res.next" }).getAttribute("href")).toContain(
		"page=2",
	);
	expect(screen.queryByRole("link", { name: "admin.res.previous" })).toBeNull();
});

test("searching and sorting are Inertia visits that keep the filters", () => {
	render(<ResourceList {...base} q="ada" />);
	fireEvent.click(screen.getByRole("button", { name: "Email" }));
	expect(get).toHaveBeenCalledWith("/admin/r/users", {
		q: "ada",
		sort: "email",
		dir: "desc",
		page: 1,
	});
});

test("an empty result says so and does not draw a table", () => {
	render(<ResourceList {...base} rows={[]} hasNext={false} />);
	expect(screen.getByText("admin.res.noRows")).toBeTruthy();
	expect(screen.queryByRole("table")).toBeNull();
});

test("flagged ratings can be restored from the list", async () => {
	render(
		<ResourceList
			{...base}
			resource={{
				slug: "rating-flags",
				navKey: "ratingFlags",
				hasDetail: false,
				searchable: false,
			}}
			columns={[
				{ key: "user_id", kind: "number", sortable: false },
				{ key: "item_id", kind: "id", sortable: false },
			]}
			rows={[{ id: "", cells: [3, "00000000-0000-0000-0000-0000000000aa"] }]}
			filters={[]}
			hasNext={false}
		/>,
	);
	fireEvent.click(screen.getByRole("button", { name: "admin.res.actions.restore" }));
	await waitFor(() => expect(reload).toHaveBeenCalled());
	expect(fetchMock.mock.calls[0][0]).toBe(
		"/admin/scoring/flags/3/00000000-0000-0000-0000-0000000000aa/restore",
	);
});

test("a user record shows its fields and lets an admin ban with a reason and an end day", async () => {
	render(
		<ResourceShow
			resource={{ slug: "users", navKey: "users" }}
			id="8"
			columns={columns}
			cells={[8, "bob@example.com", "user", null]}
		/>,
	);
	expect(screen.getByText("bob@example.com")).toBeTruthy();
	fireEvent.change(screen.getByLabelText("admin.res.actions.banReason"), {
		target: { value: "spam" },
	});
	fireEvent.change(screen.getByLabelText("admin.res.actions.banUntil"), {
		target: { value: "2026-12-31" },
	});
	fireEvent.click(screen.getByRole("button", { name: "admin.res.actions.ban" }));
	await waitFor(() => expect(fetchMock).toHaveBeenCalled());
	const body = fetchMock.mock.calls[0][1].body as URLSearchParams;
	expect(fetchMock.mock.calls[0][0]).toBe("/admin/users/8/ban");
	expect(body.get("reason")).toBe("spam");
	expect(body.get("expires_at")).toBe("2026-12-31T23:59:59Z");
	expect(body.get("expires_day")).toBeNull();
});

test("a banned user offers to lift the ban, and an open report offers to resolve", () => {
	const { rerender } = render(
		<ResourceShow
			resource={{ slug: "users", navKey: "users" }}
			id="8"
			columns={columns}
			cells={[8, "bob@example.com", "user", "2026-11-01T10:00:00Z"]}
		/>,
	);
	expect(screen.getByRole("button", { name: "admin.res.actions.unban" })).toBeTruthy();
	rerender(
		<ResourceShow
			resource={{ slug: "reports", navKey: "reports" }}
			id="R1"
			columns={[{ key: "status", kind: "badge", sortable: false }]}
			cells={["open"]}
		/>,
	);
	expect(screen.getByRole("button", { name: "admin.res.actions.resolve" })).toBeTruthy();
	expect(screen.getByRole("option", { name: "admin.res.outcomes.user_banned" })).toBeTruthy();
});

test("the flags screen lists values and marks locked and overridden flags", () => {
	render(
		<Flags
			flags={[
				{
					name: "ingest.enabled",
					description: "Master switch",
					kind: "bool",
					default: "false",
					value: "true",
					locked: false,
					stored: true,
				},
				{
					name: "social",
					description: "Phase 7",
					kind: "bool",
					default: "false",
					value: "false",
					locked: true,
					stored: false,
				},
			]}
		/>,
	);
	expect(screen.getByText("admin.flags.overridden")).toBeTruthy();
	expect(screen.getByText("admin.flags.locked")).toBeTruthy();
	expect(screen.getByRole("link", { name: "admin.flags.change" }).getAttribute("href")).toBe(
		"/admin/_flags",
	);
});
