import { createVerifiedAccount, open } from "../helpers/account";
import { expectNoSeriousA11yViolations } from "../helpers/axe";
import { expect, test } from "../helpers/fixtures";
import { canReadMail } from "../helpers/mail";

test.describe("api keys", () => {
	test("API keys page requires login", async ({ page, prefix }) => {
		await page.goto(`${prefix}/profile/api`);
		await expect(page).toHaveURL(/\/login/);
	});

	test.describe("with a verified account", () => {
		test.skip(
			!canReadMail(),
			"needs docker access to the app container logs (E2E_APP_CONTAINER)",
		);

		test("a key is shown once, listed by prefix, then revoked", async ({ page, prefix, t }) => {
			await createVerifiedAccount(page, prefix, t);
			await open(page, `${prefix}/profile/api`);
			await expect(page.getByText(t("api.profile.keys.empty"))).toBeVisible();
			await expectNoSeriousA11yViolations(page);

			await page.getByLabel(t("api.profile.keys.newName")).fill("e2e key");
			await page.getByRole("button", { name: t("api.profile.keys.create") }).click();
			const secret = page.locator("section[aria-labelledby=new-key-h] code");
			await expect(secret).toBeVisible();
			const full = ((await secret.textContent()) ?? "").trim();
			await expectNoSeriousA11yViolations(page);

			await page.getByRole("button", { name: t("api.profile.keys.dismiss") }).click();
			await expect(secret).toBeHidden();
			await page.reload();
			await expect(page.getByText(full, { exact: true })).toHaveCount(0);
			const row = page.getByRole("row", { name: /e2e key/ });
			await expect(row.getByText(t("api.profile.keys.active"))).toBeVisible();

			await row.getByRole("button", { name: t("api.profile.keys.revoke") }).click();
			await row
				.getByRole("group")
				.getByRole("button", { name: t("api.profile.keys.revoke") })
				.click();
			await expect(row.getByText(t("api.profile.keys.revoked"))).toBeVisible();
		});
	});
});
