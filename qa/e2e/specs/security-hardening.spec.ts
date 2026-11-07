import { createVerifiedAccount } from "../helpers/account";
import { expect, test } from "../helpers/fixtures";
import { canReadMail } from "../helpers/mail";
import { translator } from "../helpers/t";

const STACK_TRACE = /goroutine \d+|panic:|\.go:\d+|runtime\/debug|stack trace/i;

test.describe("security hardening", () => {
	test("error responses never leak stack traces or Go internals", async ({ request }) => {
		const probes = [
			"/games/%ZZ",
			"/api/items/%00/ratings",
			"/api/v1/games/%E0%A4%A",
			"/app/v1/items/%25/related",
			"/docs/%2e%2e%2fetc%2fpasswd",
		];
		for (const path of probes) {
			const response = await request.get(path);
			expect(response.status(), path).toBeGreaterThanOrEqual(400);
			expect(await response.text(), path).not.toMatch(STACK_TRACE);
		}
	});

	test("HTML responses are not sniffed and send a referrer policy", async ({ request }) => {
		const response = await request.get("/");
		const headers = response.headers();
		expect(headers["x-content-type-options"]).toBe("nosniff");
		expect(headers["referrer-policy"], "referrer-policy header").toBeTruthy();
	});

	test.describe("csrf", () => {
		test.skip(
			!canReadMail(),
			"needs docker access to seed a verified account (E2E_APP_CONTAINER)",
		);

		test("a state-changing request from a signed-in user without a CSRF token is rejected", async ({
			page,
			prefix,
			locale,
		}) => {
			const t = translator(locale);
			await createVerifiedAccount(page, prefix, t);

			const response = await page.request.patch(`${prefix}/profile/privacy`, {
				headers: {
					accept: "application/json",
					"content-type": "application/json",
					"sec-fetch-site": "cross-site",
					origin: "https://evil.example",
				},
				data: {},
			});
			expect(response.status()).toBe(403);
			expect(response.headers()["content-type"] ?? "").toContain("application/problem+json");
		});
	});
});
