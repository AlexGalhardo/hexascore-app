import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { expect, test, vi } from "vitest";

import { PasskeySignIn } from "./Login";

vi.mock("../../lib/useLocalized", () => ({ useLocalized: () => (path: string) => path }));
vi.mock("../../lib/useT", () => ({
	useT: () => (key: string) =>
		key === "auth.errors.passkeyFailed"
			? "Passkey sign-in was cancelled or failed."
			: key === "auth.passkey"
				? "Sign in with a passkey"
				: key,
}));

test("a cancelled passkey prompt shows the failure", async () => {
	vi.stubGlobal("PublicKeyCredential", {
		parseRequestOptionsFromJSON(options: unknown): unknown {
			return options;
		},
	});

	const get = vi
		.fn()
		.mockRejectedValue(new DOMException("The operation was cancelled.", "NotAllowedError"));
	Object.defineProperty(navigator, "credentials", { configurable: true, value: { get } });
	vi.stubGlobal(
		"fetch",
		vi.fn().mockResolvedValue({
			ok: true,
			json: async () => ({ challenge: "x" }),
		}),
	);

	render(<PasskeySignIn />);
	await userEvent.click(screen.getByRole("button", { name: "Sign in with a passkey" }));

	expect((await screen.findByRole("alert")).textContent).toContain(
		"Passkey sign-in was cancelled or failed.",
	);
	expect(get).toHaveBeenCalled();
});
