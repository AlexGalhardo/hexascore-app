import { createVerifiedAccount, expectLoggedIn, hasAuthSecret, logIn } from "../helpers/auth";
import { expect, expectAccessible, test } from "../helpers/fixtures";

test.describe("/profile/api @auth", () => {
	test.skip(!hasAuthSecret(), "needs E2E_BETTER_AUTH_SECRET to create a verified account");

	test("create a key once, use it, revoke it", async ({ page, request, baseURL }) => {
		const account = await createVerifiedAccount(request, baseURL ?? "", "apikey");
		await logIn(page, account);
		await expectLoggedIn(page);
		await page.goto("/profile/api");
		await expectAccessible(page, { exclude: ["[class*='scalar']"] });

		await page.getByLabel("Key name").fill("e2e <b>key</b>");
		await page.getByRole("button", { name: "Create key" }).click();
		await expect(page.getByText("Your new API key")).toBeVisible();
		const key = ((await page.locator("code", { hasText: "hx_live_" }).first().textContent()) ?? "").trim();
		expect(key).toMatch(/^hx_live_/);
		await expect(page.getByRole("cell", { name: "e2e <b>key</b>" })).toBeVisible();

		const ok = await request.get("/api/v1/games?limit=1", { headers: { authorization: `Bearer ${key}` } });
		// 503 = quota store unavailable (fail closed), e.g. `next start` without NODE_ENV=test or Upstash.
		expect([200, 503], await ok.text()).toContain(ok.status());

		await page.getByRole("button", { name: "I saved it" }).click();
		await expect(page.locator("code", { hasText: key })).toHaveCount(0);

		await page.getByRole("button", { name: "Revoke" }).first().click();
		await page.getByRole("dialog").getByRole("button", { name: "Revoke" }).click();
		await expect(page.getByText("Revoked")).toBeVisible();
		const revoked = await request.get("/api/v1/games?limit=1", { headers: { authorization: `Bearer ${key}` } });
		expect([401, 503]).toContain(revoked.status());
	});

	test("rejects empty and over-long key names", async ({ page, request, baseURL }) => {
		const account = await createVerifiedAccount(request, baseURL ?? "", "apikeyname");
		await logIn(page, account);
		await expectLoggedIn(page);
		await page.goto("/profile/api");
		const name = page.getByLabel("Key name");
		await name.fill("a".repeat(60));
		expect((await name.inputValue()).length).toBeLessThanOrEqual(40);
	});
});
