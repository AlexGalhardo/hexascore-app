import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, expect, test, vi } from "vitest";

import NewsletterForm from "./NewsletterForm";

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
});

test("a network failure shows the generic error", async () => {
	fetchMock.mockRejectedValue(new TypeError("offline"));
	submit();

	await waitFor(() =>
		expect(screen.getByRole("alert").textContent).toBe("footer.newsletterError"),
	);
});
