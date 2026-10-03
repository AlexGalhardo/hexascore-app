import {
	createVerifiedAccount,
	expectLoggedIn,
	hasAuthSecret,
	logIn,
	newAccount,
	verificationUrl,
} from "../helpers/auth";
import { findUser, hasDatabase } from "../helpers/db";
import { expect, expectAccessible, test } from "../helpers/fixtures";
import { enableTwoFactor, logOut } from "../helpers/session";
import { freshTotp } from "../helpers/totp";

test.describe("auth @auth", () => {
	test.describe.configure({ mode: "serial" });
	test.skip(
		!hasAuthSecret(),
		"needs E2E_BETTER_AUTH_SECRET (the deployment's BETTER_AUTH_SECRET) to follow email links",
	);

	test("signup → verification link → logged in", async ({ page }) => {
		const account = newAccount("signup");
		await page.goto("/signup");
		await page.getByLabel("Nickname").fill(account.nickname);
		await page.getByLabel("Name", { exact: true }).fill(account.name);
		await page.getByLabel("Email").fill(account.email);
		await page.locator("#password").fill(account.password);
		await expectAccessible(page);
		// Parallel workers share one IP and Better Auth's sign-up window (createVerifiedAccount in other specs).
		for (let attempt = 0; attempt < 4; attempt++) {
			const response = page.waitForResponse((r) => r.url().includes("/api/auth/sign-up/email"));
			await page.getByRole("button", { name: "Create account" }).click();
			if ((await response).status() !== 429) break;
			test.info().setTimeout(test.info().timeout + 12_000);
			await page.waitForTimeout(11_000);
		}
		await expect(page.getByText("Check your email")).toBeVisible();
		await expect(page.getByText(account.email)).toBeVisible();

		if (hasDatabase()) expect((await findUser(account.email))?.emailVerified).toBe(false);
		await page.goto(verificationUrl(account.email));
		await expect(page.getByRole("heading", { name: "Email confirmed" })).toBeVisible();
		await expectLoggedIn(page);
		if (hasDatabase()) expect((await findUser(account.email))?.emailVerified).toBe(true);
	});

	test("signup validates nickname and password rules on the client", async ({ page }) => {
		await page.goto("/signup");
		await page.getByLabel("Nickname").fill("ab");
		await page.getByLabel("Name", { exact: true }).fill("x");
		await page.getByLabel("Email").fill("not-an-email");
		await page.locator("#password").fill("short");
		await page.getByRole("button", { name: "Create account" }).click();
		await expect(page.getByLabel("Nickname")).toHaveAttribute("aria-invalid", "true");
		await expect(page.locator("#password")).toHaveAttribute("aria-invalid", "true");
		await expect(page.getByText("Check your email")).toBeHidden();
	});

	test("wrong password shows a generic error", async ({ page, request, baseURL }) => {
		const account = await createVerifiedAccount(request, baseURL ?? "", "badpw");
		await logIn(page, { ...account, password: "Wrong-Passw0rd!" });
		await expect(page.getByText(/Wrong email\/nickname or password|Too many attempts/)).toBeVisible();
	});

	test("login with TOTP two-factor", async ({ page, request, baseURL }) => {
		const account = await createVerifiedAccount(request, baseURL ?? "", "2fa");
		await logIn(page, account);
		await expectLoggedIn(page);
		const secret = await enableTwoFactor(page, account.password, freshTotp);
		if (hasDatabase()) expect((await findUser(account.email))?.twoFactorEnabled).toBe(true);

		await logOut(page);
		await logIn(page, account);
		await expect(page.getByRole("heading", { name: "Two-factor authentication" })).toBeVisible();
		await page.getByLabel("Code").fill("000000");
		await page.getByRole("button", { name: "Verify" }).click();
		await expect(page.getByText("Invalid code. Try again.")).toBeVisible();
		await page.getByLabel("Code").fill(await freshTotp(secret));
		await page.getByRole("button", { name: "Verify" }).click();
		await expectLoggedIn(page);
	});
});
