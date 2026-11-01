import { cleanup, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, expect, test, vi } from "vitest";

import ReportDialog from "./ReportDialog";

vi.mock("@inertiajs/react", () => ({
	useForm: (data: Record<string, unknown>) => ({
		data,
		setData: vi.fn(),
		post: vi.fn(),
		reset: vi.fn(),
		processing: false,
		errors: {},
	}),
}));
vi.mock("../lib/useLocalized", () => ({ useLocalized: () => (path: string) => path }));
vi.mock("../lib/useT", () => ({ useT: () => (key: string) => key }));

afterEach(cleanup);

test("the report dialog loads on the first click and Escape returns focus to the button", async () => {
	const user = userEvent.setup();
	render(<ReportDialog targetType="review" targetId="r1" />);
	expect(screen.queryByRole("dialog")).toBeNull();
	const button = screen.getByRole("button", { name: "social.report.reportReview" });
	await user.click(button);
	const dialog = await screen.findByRole("dialog", {}, { timeout: 5000 });
	expect(dialog.querySelector("select")).not.toBeNull();
	await user.keyboard("{Escape}");
	expect(screen.queryByRole("dialog")).toBeNull();
	expect(document.activeElement).toBe(button);
});
