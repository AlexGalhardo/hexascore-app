import { afterEach, beforeEach, expect, test, vi } from "vitest";

import { cleanup, fireEvent, render, screen, waitFor } from "../../lib/testing";
import Locks from "./Locks";

const reload = vi.fn();

vi.mock("@inertiajs/react", () => ({
	Head: () => null,
	Link: ({ href, children }: { href: string; children: React.ReactNode }) => (
		<a href={href}>{children}</a>
	),
	router: { reload: () => reload() },
}));
vi.mock("../../lib/useT", () => ({
	useT: () => (key: string, p?: Record<string, string>) =>
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

const lock = {
	id: "L1",
	targetType: "item",
	targetId: "00000000-0000-0000-0000-0000000000aa",
	field: "title",
	reason: "edit war <b>",
	lockedBy: 3,
	expiresAt: "",
};

test("lists a lock as text and a lock without expiry says never", () => {
	render(<Locks locks={[lock]} targets={["item", "edition"]} />);
	expect(screen.getByText("edit war <b>")).toBeTruthy();
	expect(screen.getByText("admin.common.never")).toBeTruthy();
	expect(screen.getByText("#3")).toBeTruthy();
});

test("unlocking posts to the delete endpoint and reloads", async () => {
	render(<Locks locks={[lock]} targets={["item"]} />);
	fireEvent.click(screen.getByRole("button", { name: "admin.locks.unlock" }));
	await waitFor(() => expect(reload).toHaveBeenCalled());
	expect(fetchMock.mock.calls[0][0]).toBe("/admin/locks/L1/delete");
});

test("the lock form posts the target, id, field and reason", async () => {
	render(<Locks locks={null} targets={["item"]} />);
	expect(screen.getByText("admin.locks.empty")).toBeTruthy();
	fireEvent.change(screen.getByLabelText("admin.locks.targetId"), {
		target: { value: "00000000-0000-0000-0000-0000000000aa" },
	});
	fireEvent.change(screen.getByLabelText("admin.locks.reason"), { target: { value: "spam" } });
	fireEvent.click(screen.getByRole("button", { name: "admin.locks.lock" }));
	await waitFor(() => expect(fetchMock).toHaveBeenCalled());
	const body = fetchMock.mock.calls[0][1].body as URLSearchParams;
	expect(fetchMock.mock.calls[0][0]).toBe("/admin/locks");
	expect(body.get("target")).toBe("item");
	expect(body.get("target_id")).toBe("00000000-0000-0000-0000-0000000000aa");
	expect(body.get("reason")).toBe("spam");
});
