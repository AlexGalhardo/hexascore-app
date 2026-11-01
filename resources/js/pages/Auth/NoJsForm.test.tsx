import { render } from "@testing-library/react";
import { expect, test, vi } from "vitest";

import Login from "./Login";
import Signup from "./Signup";

vi.mock("@inertiajs/react", () => ({
	Head: () => null,
	usePage: () => ({ props: {} }),
	Link: ({ children }: { children: React.ReactNode }) => <a href="/">{children}</a>,
	useForm: <T extends object>(data: T) => ({
		data,
		errors: {},
		processing: false,
		setData: () => undefined,
		post: () => undefined,
		reset: () => undefined,
	}),
}));
vi.mock("../../lib/useLocalized", () => ({ useLocalized: () => (path: string) => path }));
vi.mock("../../lib/useT", () => ({ useT: () => (key: string) => key }));

const names = (form: HTMLFormElement): string[] =>
	[...form.elements].map((el) => el.getAttribute("name") ?? "").filter(Boolean);

// A browser without scripts posts only the named controls; a missing name or form_token
// would sign the user up into the honeypot's silent drop.
test("signup form posts every field the server reads, including the form token", () => {
	const { container } = render(<Signup formToken="tok" />);
	const form = container.querySelector("form") as HTMLFormElement;

	expect(names(form)).toEqual(
		expect.arrayContaining([
			"name",
			"nickname",
			"email",
			"password",
			"password_confirmation",
			"website",
			"form_token",
		]),
	);
	expect((form.elements.namedItem("form_token") as HTMLInputElement).value).toBe("tok");
});

test("login form posts identifier, password and remember", () => {
	const { container } = render(<Login />);

	expect(names(container.querySelector("form") as HTMLFormElement)).toEqual(
		expect.arrayContaining(["identifier", "password", "remember"]),
	);
});

test("signup enforces the nickname and email bounds and explains the nickname is permanent", () => {
	const { container } = render(<Signup formToken="tok" />);
	const form = container.querySelector("form") as HTMLFormElement;
	const nickname = form.elements.namedItem("nickname") as HTMLInputElement;
	const email = form.elements.namedItem("email") as HTMLInputElement;

	expect([nickname.minLength, nickname.maxLength, nickname.pattern, nickname.required]).toEqual([
		4,
		16,
		"[a-z]{4,16}",
		true,
	]);
	expect(
		document.getElementById(nickname.getAttribute("aria-describedby") ?? "")?.textContent,
	).toBe("auth.signup.nicknameHint");
	expect([email.minLength, email.maxLength]).toEqual([12, 48]);
});

test("the auth pages no longer show the members label", () => {
	for (const page of [<Signup key="s" formToken="tok" />, <Login key="l" />]) {
		const { container, unmount } = render(page);
		expect(container.textContent).not.toContain("auth.members");
		unmount();
	}
});
