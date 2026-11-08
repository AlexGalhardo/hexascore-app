import { cleanup, render, screen } from "@testing-library/react";
import { afterEach, expect, test, vi } from "vitest";

import Flash from "./Flash";

vi.mock("../lib/useT", () => ({ useT: () => (key: string) => key }));

afterEach(cleanup);

test.each([
	["status", "bg-green-500", "status"],
	["warning", "bg-yellow-400", "alert"],
	["error", "bg-red-500", "alert"],
] as const)("%s toast is vibrant, black, bold and uppercase", (kind, bg, role) => {
	render(<Flash message="Saved" kind={kind} />);
	const toast = screen.getByRole(role);
	for (const c of [bg, "text-black", "font-extrabold", "uppercase", "text-lg"]) {
		expect(toast.className).toContain(c);
	}
});
