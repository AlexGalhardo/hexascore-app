import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, expect, test, vi } from "vitest";

import DataExport from "./DataExport";

vi.mock("../lib/useT", () => ({ useT: () => (key: string) => key }));

afterEach(() => vi.unstubAllGlobals());

test("posts the export request and shows the cooldown answer", async () => {
	const fetchMock = vi.fn().mockResolvedValue({ ok: false, status: 429 });
	vi.stubGlobal("fetch", fetchMock);

	render(<DataExport />);
	await userEvent.click(screen.getByRole("button"));

	expect(fetchMock.mock.calls[0][0]).toBe("/profile/export");
	expect(fetchMock.mock.calls[0][1].method).toBe("POST");
	expect((await screen.findByRole("status")).textContent).toContain("account.export.cooldown");
});
