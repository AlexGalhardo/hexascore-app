import { expect, expectAccessible, LOCALES, test } from "../helpers/fixtures";

test.describe("404s", () => {
	for (const { prefix } of LOCALES) {
		for (const path of ["/games/this-slug-does-not-exist", "/no-such-section", "/docs/nope", "/a/b/c"]) {
			test(`${prefix}${path} answers 404 with the localized page`, async ({ page }) => {
				const response = await page.goto(`${prefix}${path}`);
				expect(response?.status()).toBe(404);
				await expect(page.getByRole("link", { name: /home|início|inicio/i })).toBeVisible();
				await expectAccessible(page);
			});
		}
	}

	test("API 404s are RFC 9457 problems", async ({ request }) => {
		const response = await request.get("/api/items/00000000-0000-0000-0000-000000000000/ratings");
		expect(response.status()).toBeLessThan(500);
		const v1 = await request.get("/api/v1/nope", { headers: { authorization: "Bearer hx_live_invalid" } });
		expect([401, 404]).toContain(v1.status());
		expect(v1.headers()["content-type"]).toContain("application/problem+json");
	});
});
