import userEvent from "@testing-library/user-event";
import { afterEach, expect, test, vi } from "vitest";
import { cleanup, render, screen } from "../lib/testing";

import DataExport from "./DataExport";

vi.mock("../lib/useT", () => ({ useT: () => (key: string) => key }));

afterEach(() => {
	cleanup();
	vi.unstubAllGlobals();
});

test("posts the export request and shows the cooldown answer", async () => {
	const fetchMock = vi.fn().mockResolvedValue({ ok: false, status: 429 });
	vi.stubGlobal("fetch", fetchMock);

	render(<DataExport />);
	await userEvent.click(screen.getByRole("button"));

	expect(fetchMock.mock.calls[0][0]).toBe("/profile/export");
	expect(fetchMock.mock.calls[0][1].method).toBe("POST");
	expect((await screen.findByRole("status")).textContent).toContain("account.export.cooldown");
});

// CSRF is Fetch Metadata (SDD 13 §7.5): the mutation must stay a same-origin, credentialed
// request to a relative URL so the browser sends Origin and Sec-Fetch-Site: same-origin.
test("the export mutation is a same-origin JSON request and reports success", async () => {
	const fetchMock = vi.fn().mockResolvedValue({ ok: true, status: 202 });
	vi.stubGlobal("fetch", fetchMock);

	render(<DataExport />);
	await userEvent.click(screen.getByRole("button"));

	const [url, init] = fetchMock.mock.calls[0];
	expect(url).toBe("/profile/export");
	expect(init).toMatchObject({
		method: "POST",
		credentials: "same-origin",
		headers: { Accept: "application/json" },
	});
	expect((await screen.findByRole("status")).textContent).toContain("account.export.requested");
});
