import { expectNoSeriousA11yViolations } from "../helpers/axe";
import { expect, test } from "../helpers/fixtures";

test.describe("404s", () => {
	for (const path of [
		"/games/this-slug-does-not-exist",
		"/no-such-section",
		"/docs/nope",
		"/a/b/c",
	]) {
		// D-41: a missing HTML page answers 404 with the localised error page and a link home.
		test(`${path} answers 404 with a link home`, async ({ page, request, prefix }) => {
			const response = await request.get(`${prefix}${path}`, { maxRedirects: 0 });
			expect(response.status()).toBe(404);
			await page.goto(`${prefix}${path}`);
			await expect(page.locator("h1")).toBeVisible();
			await expect(page.locator(`a[href="${prefix}/"]`).first()).toBeVisible();
			await expectNoSeriousA11yViolations(page);
		});
	}

	test("API 404s are RFC 9457 problems", async ({ request }) => {
		const v1 = await request.get("/api/v1/nope", {
			headers: { authorization: "Bearer hx_live_invalid" },
		});
		expect([401, 404]).toContain(v1.status());
		expect(v1.headers()["content-type"]).toContain("application/problem+json");
	});
});
