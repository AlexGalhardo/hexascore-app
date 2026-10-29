import { createVerifiedAccount, logIn, open, signOut } from "../helpers/account";
import { expectNoSeriousA11yViolations } from "../helpers/axe";
import { expect, test } from "../helpers/fixtures";
import { canReadMail } from "../helpers/mail";

test.describe("profile", () => {
	test("profile requires login", async ({ page, prefix }) => {
		await page.goto(`${prefix}/profile`);
		await expect(page).toHaveURL(/\/login/);
	});

	test.describe("with a verified account", () => {
		test.skip(!canReadMail(), "needs docker access to the app container logs (E2E_APP_CONTAINER)");

		test("change password, then log in with the new one", async ({ page, prefix, t, any }) => {
			const account = await createVerifiedAccount(page, prefix, any);
			await open(page, `${prefix}/profile`);
			await expectNoSeriousA11yViolations(page);

			const next = "N3w-Passw0rd!x2";
			await page.getByLabel(t("auth.fields.currentPassword")).fill(account.password);
			await page.getByLabel(t("auth.fields.newPassword")).fill(next);
			const saved = page.waitForResponse((r) => r.url().endsWith("/profile/password") && r.request().method() === "PUT");
			await page.getByRole("button", { name: t("account.password.change"), exact: true }).click();
			expect((await saved).status()).toBeLessThan(400);

			await signOut(page);
			await logIn(page, prefix, any, account.email, next);
			await page.goto(`${prefix}/profile`);
			await expect(page).toHaveURL(/\/profile$/);
		});

		test("account deletion: 10 s countdown, cancel, confirm, banner and undo", async ({ page, prefix, t, any }) => {
			test.setTimeout(120_000);
			const account = await createVerifiedAccount(page, prefix, any);
			await open(page, `${prefix}/profile`);

			await page.getByRole("button", { name: t("account.delete.button") }).click();
			const dialog = page.getByRole("dialog", { name: t("account.delete.dialogTitle") });
			const confirm = dialog.getByRole("button", { name: /^(Confirm|Confirmar|Confirmar) \(\d+\)$/ });
			await expect(confirm).toBeDisabled();
			await expectNoSeriousA11yViolations(page);
			await dialog.getByRole("button", { name: t("account.delete.cancel") }).click();
			await expect(dialog).toBeHidden();

			await page.getByRole("button", { name: t("account.delete.button") }).click();
			const ready = dialog.getByRole("button", { name: t("account.delete.confirm") });
			await expect(ready).toBeEnabled({ timeout: 15_000 });
			await ready.click();
			await page.waitForURL(/\/login/);

			await logIn(page, prefix, any, account.email, account.password);
			const banner = page.getByRole("button", { name: any("account.banner.cancel") });
			await expect(banner).toBeVisible();
			await banner.click();
			await expect(banner).toBeHidden();
		});
	});
});
