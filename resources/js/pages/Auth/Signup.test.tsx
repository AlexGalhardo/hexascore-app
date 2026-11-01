import { render, screen } from "@testing-library/react";
import { expect, test, vi } from "vitest";

import { GoogleButton } from "../../components/auth/AuthFrame";
import { Agreement } from "./Signup";

vi.mock("../../lib/useT", () => ({ useT: () => (key: string) => key }));

test("the agreement links the terms and the privacy policy", () => {
	render(
		<Agreement
			text="By creating an account you agree to the <terms>Terms</terms> and the <privacy>Privacy policy</privacy>."
			l={(path) => `/es${path}`}
		/>,
	);

	expect(screen.getByRole("link", { name: "Terms" }).getAttribute("href")).toBe("/es/terms");
	expect(screen.getByRole("link", { name: "Privacy policy" }).getAttribute("href")).toBe(
		"/es/policy",
	);
	expect(screen.getByText(/By creating an account/).textContent).toBe(
		"By creating an account you agree to the Terms and the Privacy policy.",
	);
});

test("the Google button is a full navigation to the OAuth redirect", () => {
	render(<GoogleButton mode="signUp" />);

	const link = screen.getByRole("link", { name: "auth.googleSignUp" });
	expect(link.getAttribute("href")).toBe("/auth/google/redirect");
});
