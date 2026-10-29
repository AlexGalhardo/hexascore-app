import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { expect, test } from "vitest";

import Field from "./Field";

test("labels the input and shows its error", async () => {
	render(<Field label="Email" error="Required" />);

	const input = screen.getByLabelText(/Email/);
	await userEvent.type(input, "ana@example.com");

	expect((input as HTMLInputElement).value).toBe("ana@example.com");
	expect(screen.getByText("Required")).toBeTruthy();
});
