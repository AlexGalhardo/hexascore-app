import { cleanup, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, expect, test, vi } from "vitest";

import ShareButton from "./ShareButton";

vi.mock("../../lib/useT", () => ({ useT: () => (key: string) => key }));

afterEach(() => {
	cleanup();
	vi.unstubAllGlobals();
});

test("uses the Web Share API when the browser has it", async () => {
	const share = vi.fn().mockResolvedValue(undefined);
	vi.stubGlobal("navigator", { share });

	render(<ShareButton title="Hades" />);
	await userEvent.click(screen.getByRole("button", { name: "social.share.action" }));

	expect(share).toHaveBeenCalledWith({ title: "Hades", url: window.location.href });
});

test("copies the link and says so when the browser cannot share", async () => {
	const writeText = vi.fn().mockResolvedValue(undefined);
	vi.stubGlobal("navigator", { clipboard: { writeText } });

	render(<ShareButton title="Hades" />);
	await userEvent.click(screen.getByRole("button", { name: "social.share.action" }));

	expect(writeText).toHaveBeenCalledWith(window.location.href);
	expect((await screen.findByRole("status")).textContent).toBe("social.share.copied");
});
