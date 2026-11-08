import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, expect, test, vi } from "vitest";

import NewsletterForm, { checkEmail } from "./NewsletterForm";

vi.mock("../lib/useT", () => ({ useT: () => (key: string) => key }));

const fetchMock = vi.fn();
vi.stubGlobal("fetch", fetchMock);

afterEach(() => {
	cleanup();
	fetchMock.mockReset();
});

const submit = () => {
	render(<NewsletterForm endpoint="https://email.test/subscribe/hexascore" />);
	fireEvent.change(screen.getByLabelText("footer.newsletterPlaceholder"), {
		target: { value: "ana@example.com" },
	});
	fireEvent.click(screen.getByRole("button", { name: "footer.newsletterSubmit" }));
};

test("posts the email and the empty honeypot without credentials, then asks to check the inbox", async () => {
	fetchMock.mockResolvedValue({ status: 202 });
	submit();

	await waitFor(() =>
		expect(screen.getByRole("status").textContent).toBe("footer.newsletterDone"),
	);
	expect(screen.getByRole("status").className).toContain("bg-green-500");
	const [url, init] = fetchMock.mock.calls[0];
	expect(url).toBe("https://email.test/subscribe/hexascore");
	expect(init.credentials).toBeUndefined();
	expect(JSON.parse(init.body)).toEqual({ email: "ana@example.com", website: "" });
});

test.each([
	[422, "footer.newsletterInvalid"],
	[429, "footer.newsletterRateLimited"],
	[500, "footer.newsletterError"],
])("status %i shows an alert", async (status, key) => {
	fetchMock.mockResolvedValue({ status });
	submit();

	await waitFor(() => expect(screen.getByRole("alert").textContent).toBe(key));
	expect(screen.getByRole("alert").className).toContain("bg-red-500");
});

test("a network failure shows the generic error", async () => {
	fetchMock.mockRejectedValue(new TypeError("offline"));
	submit();

	await waitFor(() =>
		expect(screen.getByRole("alert").textContent).toBe("footer.newsletterError"),
	);
});

test("checkEmail keeps addresses to 12..48 characters with a real domain", () => {
	expect(checkEmail("ana@example.com")).toBeNull();
	expect(checkEmail("a@b.co")).toBe("length"); // 6 characters
	expect(checkEmail(`${"a".repeat(40)}@example.com`)).toBe("length"); // 52 characters
	expect(checkEmail("ana.example.com")).toBe("invalid");
	expect(checkEmail("ana@examplecom")).toBe("invalid");
	expect(checkEmail("ana @example.com")).toBe("invalid");
	expect(checkEmail("ana@example.c")).toBe("invalid");
});

test("does not post an address outside the length rule", async () => {
	render(<NewsletterForm endpoint="https://email.test/subscribe/hexascore" />);
	fireEvent.change(screen.getByLabelText("footer.newsletterPlaceholder"), {
		target: { value: "a@b.co" },
	});
	fireEvent.click(screen.getByRole("button", { name: "footer.newsletterSubmit" }));

	await waitFor(() =>
		expect(screen.getByRole("alert").textContent).toBe("footer.newsletterLength"),
	);
	expect(fetchMock).not.toHaveBeenCalled();
});
