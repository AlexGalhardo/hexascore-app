import { expect, type Page } from "@playwright/test";
import { mailLink } from "./mail";
import { anyLocale } from "./t";

export interface Account {
	name: string;
	email: string;
	nickname: string;
	password: string;
}

// signup, login and reset demand a 21-bit proof of work from the browser (GET /challenge), which takes seconds.
export const POW_TIMEOUT = 60_000;

const LETTERS = "abcdefghijklmnopqrstuvwxyz";

function randomLetters(length: number): string {
	return Array.from({ length }, () => LETTERS[Math.floor(Math.random() * LETTERS.length)]).join(
		"",
	);
}

// Nicknames are 4 to 16 lowercase letters, so uniqueness comes from random letters rather than digits.
export function newAccount(): Account {
	const nickname = `e${randomLetters(10)}`;
	return {
		name: `E2E ${nickname}`,
		email: `${nickname}@example.test`,
		nickname,
		password: "E2e-Passw0rd!x1",
	};
}

// Server-rendered forms submit natively (a GET that reloads the page) when clicked before React hydrates,
// so wait until the first form carries React props. Pages without a form only need the load event.
export async function open(page: Page, url: string): Promise<void> {
	await page.goto(url);
	await page.waitForFunction(() => {
		const form = document.querySelector("main form");
		return !form || Object.keys(form).some((key) => key.startsWith("__reactProps$"));
	});
}

type Translate = (key: string) => string | RegExp;

export async function fillSignup(
	page: Page,
	prefix: string,
	t: Translate,
	account: Account,
): Promise<void> {
	await open(page, `${prefix}/signup`);
	// Inputs intermittently lose a value filled right after hydration on narrow viewports; retry until each sticks.
	const fields = [
		[page.getByLabel(t("auth.fields.name")).first(), account.name],
		[page.getByLabel(t("auth.fields.email")), account.email],
		[page.locator("input[name=password]"), account.password],
		[page.locator("input[name=password_confirmation]"), account.password],
	] as const;
	for (const [input, value] of fields) {
		await expect(async () => {
			await input.fill(value);
			await expect(input).toHaveValue(value, { timeout: 1_000 });
		}).toPass();
	}
}

// Waits for the post to settle (unless a failure is expected) so a following navigation cannot cancel it mid-request.
export async function logIn(
	page: Page,
	prefix: string,
	t: Translate,
	identifier: string,
	password: string,
	{ expectFailure = false } = {},
): Promise<void> {
	if (!/\/login$/.test(new URL(page.url(), "http://x").pathname))
		await open(page, `${prefix}/login`);
	await page.getByLabel(t("auth.fields.identifier")).fill(identifier);
	await page.locator("input[name=password]").fill(password);
	await page.getByRole("button", { name: t("auth.login.submit") }).click();
	if (!expectFailure)
		await page.waitForURL((url) => !/\/login$/.test(url.pathname), { timeout: POW_TIMEOUT });
}

// Signs up through the UI, follows the mailed link (the guard sends a guest to /login and back to it),
// then picks the nickname on the onboarding page; the account ends up verified and signed in.
export async function createVerifiedAccount(
	page: Page,
	prefix: string,
	t: Translate,
): Promise<Account> {
	const account = newAccount();
	await fillSignup(page, prefix, t, account);
	// A click right after hydration is intermittently dropped on the signup form; click again until the post goes out.
	await expect(async () => {
		await page.getByRole("button", { name: t("auth.signup.submit") }).click();
		await expect(page).toHaveURL(/\/signup\/check-email$/, { timeout: 20_000 });
	}).toPass({ timeout: POW_TIMEOUT * 2 });
	const link = await mailLink(account.email, "/verify-email/confirm/");
	await open(page, link);
	await expect(page).toHaveURL(/\/login/);
	await page.getByLabel(anyLocale("auth.fields.identifier")).fill(account.email);
	await page.locator("input[name=password]").fill(account.password);
	await page.getByRole("button", { name: anyLocale("auth.login.submit") }).click();
	await completeOnboarding(page, anyLocale, account);
	return account;
}

export async function completeOnboarding(
	page: Page,
	t: Translate,
	account: Account,
): Promise<void> {
	await page.waitForURL((url) => !/\/(login|signup)/.test(url.pathname), {
		timeout: POW_TIMEOUT,
	});
	// Login lands on the dashboard; the nickname form is only shown when asked for.
	await page.goto("/onboarding");
	const nickname = page.getByLabel(t("auth.fields.nickname"));
	if (
		await nickname.waitFor({ timeout: 5_000 }).then(
			() => true,
			() => false,
		)
	) {
		await nickname.fill(account.nickname);
		await page.getByRole("button", { name: t("auth.onboarding.submit") }).click();
		await page.waitForURL((url) => !url.pathname.includes("onboarding"));
	}
}

// Dropping the cookies is the fastest sign-out; the logout button itself is covered by the navigation specs.
// It settles in-flight visits first: a late response re-issues the session cookie and undoes the logout.
export async function signOut(page: Page): Promise<void> {
	await page.waitForLoadState("networkidle");
	await page.context().clearCookies();
}

export async function expectSignedIn(page: Page, prefix: string): Promise<void> {
	await page.goto(`${prefix}/profile`);
	await expect(page).toHaveURL(/\/profile$/);
}
