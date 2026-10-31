import { render } from "@testing-library/react";
import { expect, test, vi } from "vitest";

import Login from "./Login";
import Signup from "./Signup";

vi.mock("@inertiajs/react", () => ({
	Head: () => null,
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
