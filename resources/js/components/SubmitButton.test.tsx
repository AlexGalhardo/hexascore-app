import { cleanup, render, screen } from "@testing-library/react";
import { afterEach, expect, test } from "vitest";

import SubmitButton from "./SubmitButton";

afterEach(cleanup);

test("is enabled and idle by default", () => {
	render(<SubmitButton>Log in</SubmitButton>);
	const b = screen.getByRole("button", { name: "Log in" }) as HTMLButtonElement;
	expect(b.disabled).toBe(false);
	expect(b.getAttribute("aria-busy")).toBe("false");
	expect(b.hasAttribute("data-pending")).toBe(false);
});

test("pending disables the button and marks it busy", () => {
	render(<SubmitButton pending>Log in</SubmitButton>);
	const b = screen.getByRole("button", { name: "Log in" }) as HTMLButtonElement;
	expect(b.disabled).toBe(true);
	expect(b.getAttribute("aria-busy")).toBe("true");
	expect(b.hasAttribute("data-pending")).toBe(true);
});
