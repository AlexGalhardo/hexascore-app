import { createVerifiedAccount, expectLoggedIn, hasAuthSecret, logIn } from "../helpers/auth";
import { findUser, hasDatabase } from "../helpers/db";
import { expect, expectAccessible, test } from "../helpers/fixtures";
import { logOut } from "../helpers/session";

test.describe("profile @auth", () => {
	test.skip(!hasAuthSecret(), "needs E2E_BETTER_AUTH_SECRET to create a verified account");

	test("profile requires login", async ({ page }) => {
		await page.goto("/profile");
		await expect(page).toHaveURL(/\/login/);
	});

	test("change password, then log in with the new one", async ({ page, request, baseURL }) => {
		const account = await createVerifiedAccount(request, baseURL ?? "", "chpw");
		await logIn(page, account);
		await expectLoggedIn(page);
		await page.goto("/profile");
		await expectAccessible(page);

		const next = "N3w-Passw0rd!x";
		await page.getByLabel("Current password").fill(account.password);
		await page.getByLabel("New password").fill(next);
		await page.getByRole("button", { name: "Change password" }).click();
		await expect(page.getByText("Password updated.")).toBeVisible();

		await logOut(page);
		await logIn(page, { ...account, password: next });
		await expectLoggedIn(page);
	});

	test("account deletion: 10 s countdown, cancel, confirm, banner and undo", async ({ page, request, baseURL }) => {
		const account = await createVerifiedAccount(request, baseURL ?? "", "delete");
		await logIn(page, account);
		await expectLoggedIn(page);
		await page.goto("/profile");

		await page.getByRole("button", { name: "Delete my account" }).click();
		const dialog = page.getByRole("dialog", { name: "Delete your account?" });
		const confirm = dialog.getByRole("button", { name: /Confirm \(\d+\)|Yes, delete my account/ });
		await expect(confirm).toBeDisabled();
		await expect(confirm).toHaveText(/Confirm \((10|9)\)/);
		await expectAccessible(page);
		await dialog.getByRole("button", { name: "Keep my account" }).click();
		await expect(dialog).toBeHidden();

		await page.getByRole("button", { name: "Delete my account" }).click();
		await expect(confirm).toHaveText("Yes, delete my account", { timeout: 15_000 });
		await expect(confirm).toBeEnabled();
		await confirm.click();
		// On mobile the navbar link sits inside the closed hamburger menu: attached, not visible.
		await expect(page.getByRole("link", { name: "Log in" }).first()).toBeAttached();
		if (hasDatabase()) expect((await findUser(account.email))?.deletionScheduledAt).not.toBeNull();

		await logIn(page, account);
		await expect(page.getByText(/scheduled for deletion/)).toBeVisible();
		await page.getByRole("button", { name: "Cancel deletion" }).click();
		await expect(page.getByText(/scheduled for deletion/)).toBeHidden();
		if (hasDatabase()) expect((await findUser(account.email))?.deletionScheduledAt).toBeNull();
	});
});
