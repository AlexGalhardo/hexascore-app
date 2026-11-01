import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { useState } from "react";
import { beforeEach, expect, test, vi } from "vitest";

import type { User } from "../../types";
import Contact from "./Index";

let user: User | null = null;

vi.mock("@inertiajs/react", () => ({
	Head: () => null,
	usePage: () => ({ props: { auth: { user } } }),
	useForm: <T extends Record<string, unknown>>(initial: T) => {
		const [data, setDataState] = useState(initial);
		return {
			data,
			errors: {},
			processing: false,
			setData: (key: keyof T, value: unknown) =>
				setDataState((d) => ({ ...d, [key]: value })),
			post: () => undefined,
			reset: () => undefined,
		};
	},
}));
vi.mock("../../lib/useLocalized", () => ({ useLocalized: () => (path: string) => path }));
vi.mock("../../lib/useT", () => ({ useT: () => (key: string) => key }));

beforeEach(() => {
	cleanup();
	user = null;
});

const input = (name: string): HTMLInputElement =>
	document.querySelector(`[name="${name}"]`) as HTMLInputElement;

test("a guest types name and email; the message has a live counter up to 512", () => {
	render(<Contact topics={["question"]} formToken="tok" />);

	expect(input("name").readOnly).toBe(false);
	expect(input("email").readOnly).toBe(false);
	expect(screen.getByTestId("contact-message-count").textContent).toBe("0/512");

	fireEvent.change(input("message"), { target: { value: "héllo" } });
	expect(screen.getByTestId("contact-message-count").textContent).toBe("5/512");
	expect((input("message") as unknown as HTMLTextAreaElement).maxLength).toBe(512);
	expect(input("message").getAttribute("aria-describedby")).toContain("contact-message-count");
});

test("a signed-in user sees the account nickname and email, read-only", () => {
	user = {
		id: 1,
		name: "Ana Silva",
		email: "ana@example.com",
		email_verified_at: null,
		nickname: "analuiza",
	};
	render(<Contact topics={["question"]} formToken="tok" />);

	expect([input("name").value, input("name").readOnly]).toEqual(["analuiza", true]);
	expect([input("email").value, input("email").readOnly]).toEqual(["ana@example.com", true]);
	expect(screen.getByLabelText("auth.fields.nickname")).toBe(input("name"));
});

test("the honeypot stays hidden from assistive tech and the form is narrow", () => {
	const { container } = render(<Contact topics={["question"]} formToken="tok" />);

	expect(input("website").closest("[aria-hidden='true']")).not.toBeNull();
	expect((container.firstElementChild as HTMLElement).className).toContain("max-w-xl");
});
