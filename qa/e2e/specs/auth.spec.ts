import { POW_TIMEOUT, createVerifiedAccount, expectSignedIn, fillSignup, logIn, newAccount, open, signOut } from "../helpers/account";
import { expectNoSeriousA11yViolations } from "../helpers/axe";
import { expect, test } from "../helpers/fixtures";
import { canReadMail, mailLink } from "../helpers/mail";
import { codeAfter, currentStep, totp } from "../helpers/totp";

// Form posts redirect to unprefixed URLs that render in the default locale, so assertions made after a
// post use `any` (copy in any locale); pages opened by URL use `t` (the project locale).
test.describe("auth", () => {
	test("login, signup and recover pages are accessible", async ({ page, prefix }) => {
		for (const path of ["/login", "/signup", "/recover-password"]) {
			await open(page, `${prefix}${path}`);
			await expect(page.locator("main h1")).toBeVisible();
			await expectNoSeriousA11yViolations(page);
		}
	});

	test("protected pages send a guest to the login form", async ({ page, prefix }) => {
		await page.goto(`${prefix}/profile`);
		await expect(page).toHaveURL(/\/login/);
	});

	test("signup refuses a weak password and stays on the form", async ({ page, prefix, any }) => {
		await fillSignup(page, prefix, any, { ...newAccount(), password: "short" });
		await page.getByRole("button", { name: any("auth.signup.submit") }).click();
		await expect(page.locator("main p.text-red-600").first()).toBeVisible({ timeout: POW_TIMEOUT });
		await expect(page).toHaveURL(/\/signup$/);
	});

	test("wrong password shows a generic error", async ({ page, prefix, any }) => {
		await logIn(page, prefix, any, "admin@example.test", "Wrong-Passw0rd!", { expectFailure: true });
		await expect(page.locator("main p.text-red-600")).toBeVisible({ timeout: POW_TIMEOUT });
		await expect(page).toHaveURL(/\/login$/);
	});

	test.describe("with mail access", () => {
		test.skip(!canReadMail(), "needs docker access to the app container logs (E2E_APP_CONTAINER)");

		test("signup, verify the mailed link and land signed in", async ({ page, prefix, any }) => {
			const account = newAccount();
			await fillSignup(page, prefix, any, account);
			await page.getByRole("button", { name: any("auth.signup.submit") }).click();
			await expect(page).toHaveURL(/\/signup\/check-email$/, { timeout: POW_TIMEOUT });
			await expectNoSeriousA11yViolations(page);

			await open(page, await mailLink(account.email, "/verify-email/confirm/"));
			await logIn(page, prefix, any, account.email, account.password);
			await expectSignedIn(page, prefix);
		});

		test("unverified account cannot reach the dashboard", async ({ page, prefix, any }) => {
			const account = newAccount();
			await fillSignup(page, prefix, any, account);
			await page.getByRole("button", { name: any("auth.signup.submit") }).click();
			await expect(page).toHaveURL(/\/signup\/check-email$/, { timeout: POW_TIMEOUT });
			await logIn(page, prefix, any, account.email, account.password);
			await page.goto(`${prefix}/dashboard`);
			await expect(page).toHaveURL(/verify-email/);
		});

		test("TOTP two-factor: enable, then log in with a code", async ({ page, prefix, any }) => {
			test.setTimeout(240_000);
			const account = await createVerifiedAccount(page, prefix, any);

			await open(page, "/user/two-factor");
			await page.getByRole("button", { name: "Enable two-factor authentication" }).click();
			const secret = (await page.locator("code").first().innerText()).trim();
			await expectNoSeriousA11yViolations(page);
			const step = currentStep();
			await page.getByLabel("Code").fill(totp(secret, step));
			await page.getByRole("button", { name: "Confirm" }).click();
			await expect(page.getByText("Two-factor authentication is on.", { exact: false }).first()).toBeVisible();

			await signOut(page);
			await logIn(page, prefix, any, account.email, account.password);
			await expect(page).toHaveURL(/two-factor-challenge/, { timeout: POW_TIMEOUT });
			await expectNoSeriousA11yViolations(page);
			await page.locator("input[autocomplete=one-time-code]").fill("000000");
			await page.getByRole("button", { name: any("auth.login.verify") }).click();
			await expect(page.locator("main p.text-red-600")).toBeVisible({ timeout: POW_TIMEOUT });
			await page.locator("input[autocomplete=one-time-code]").fill(await codeAfter(secret, step));
			await page.getByRole("button", { name: any("auth.login.verify") }).click();
			await page.waitForURL((url) => !url.pathname.includes("two-factor-challenge"));
			await expectSignedIn(page, prefix);
		});

		test("password reset: mailed link sets a new password", async ({ page, prefix, any }) => {
			const account = await createVerifiedAccount(page, prefix, any);
			await signOut(page);

			await open(page, `${prefix}/recover-password`);
			await page.getByLabel(any("auth.fields.email")).fill(account.email);
			await page.getByRole("button", { name: any("auth.recover.submit") }).click();
			await expect(page.getByRole("status").filter({ hasText: /\S/ }).first()).toBeVisible({ timeout: POW_TIMEOUT });

			const next = "R3set-Passw0rd!y";
			await open(page, await mailLink(account.email, "/reset-password/"));
			await page.locator("input[name=password]").fill(next);
			await page.locator("input[name=password_confirmation]").fill(next);
			await expectNoSeriousA11yViolations(page);
			await page.getByRole("button", { name: any("auth.reset.submit") }).click();
			await page.waitForURL(/\/login/, { timeout: POW_TIMEOUT });

			await logIn(page, prefix, any, account.email, next);
			await expectSignedIn(page, prefix);
		});
	});
});
