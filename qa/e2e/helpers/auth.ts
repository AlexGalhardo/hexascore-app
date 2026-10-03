import { createHmac, randomBytes } from "node:crypto";
import { type APIRequestContext, expect, type Page, test } from "@playwright/test";
import { E2E_EMAIL_DOMAIN } from "./db";

export const PASSWORD = "E2e-Passw0rd!";

export type TestAccount = { email: string; nickname: string; name: string; password: string };

export function newAccount(tag: string): TestAccount {
	const id = randomBytes(4).toString("hex");
	return {
		email: `e2e-${tag}-${id}@${E2E_EMAIL_DOMAIN}`,
		nickname: `e2e_${id}`,
		name: `E2E ${tag}`,
		password: PASSWORD,
	};
}

export function hasAuthSecret(): boolean {
	return Boolean(process.env.E2E_BETTER_AUTH_SECRET);
}

function b64url(input: Buffer | string): string {
	return Buffer.from(input).toString("base64url");
}

/**
 * The link Better Auth emails on signup. Better Auth 1.7 does not store it (it is an HS256 JWT `{ email }` signed with
 * BETTER_AUTH_SECRET), so the helper mints the identical token with the deployment's secret (E2E_BETTER_AUTH_SECRET)
 * instead of reading an inbox. The app is untouched: the token goes through the real `/api/auth/verify-email`.
 */
export function verificationUrl(email: string, callbackPath = "/verify-email"): string {
	const secret = process.env.E2E_BETTER_AUTH_SECRET;
	if (!secret) throw new Error("E2E_BETTER_AUTH_SECRET is not set");
	const now = Math.floor(Date.now() / 1000);
	const header = b64url(JSON.stringify({ alg: "HS256" }));
	const payload = b64url(JSON.stringify({ email: email.toLowerCase(), iat: now, exp: now + 3600 }));
	const signature = b64url(createHmac("sha256", secret).update(`${header}.${payload}`).digest());
	return `/api/auth/verify-email?token=${header}.${payload}.${signature}&callbackURL=${encodeURIComponent(callbackPath)}`;
}

/** Better Auth answers 429 after a few rapid auth calls from one IP; wait the window out instead of failing. */
async function withRateLimitRetry<T extends { status(): number }>(call: () => Promise<T>): Promise<T> {
	for (let attempt = 0; ; attempt++) {
		const response = await call();
		if (response.status() !== 429 || attempt === 3) return response;
		await new Promise((resolve) => setTimeout(resolve, 11_000));
	}
}

/** Creates an account through the public signup endpoint (same call the form makes) and verifies it. */
export async function createVerifiedAccount(
	request: APIRequestContext,
	baseURL: string,
	tag: string,
): Promise<TestAccount> {
	const account = newAccount(tag);
	const response = await withRateLimitRetry(() =>
		request.post("/api/auth/sign-up/email", {
			headers: { origin: new URL(baseURL).origin },
			data: {
				email: account.email,
				password: account.password,
				name: account.name,
				username: account.nickname,
				callbackURL: "/verify-email",
			},
		}),
	);
	expect(response.status(), await response.text()).toBe(200);
	const verified = await request.get(verificationUrl(account.email), { maxRedirects: 0 });
	expect([200, 302]).toContain(verified.status());
	return account;
}

/** Submits the login form; waits out Better Auth's per-IP rate limit (parallel workers share one IP). */
export async function logIn(page: Page, account: TestAccount, locale = ""): Promise<void> {
	for (let attempt = 0; attempt < 6; attempt++) {
		await page.goto(`${locale}/login`);
		await page.locator("#identifier").fill(account.email);
		await page.locator("#password").fill(account.password);
		await page
			.locator("form")
			.getByRole("button", { name: /^(log in|entrar|iniciar sesión)$/i })
			.click();
		const limited = page.getByText(/Too many attempts|Muitas tentativas|Demasiados intentos/);
		const done = page
			.getByRole("button", { name: /account menu|menu da conta|menú de la cuenta/i })
			.or(page.locator("#code"))
			.or(page.getByRole("alert"));
		await expect(limited.or(done).first()).toBeVisible();
		if (!(await limited.isVisible())) return;
		// Better Auth's sign-in window can last up to a minute; extend the test instead of failing it.
		test.info().setTimeout(test.info().timeout + 25_000);
		await page.waitForTimeout(20_000);
	}
}

export async function expectLoggedIn(page: Page): Promise<void> {
	await expect(page.getByRole("button", { name: /account menu|menu da conta|menú de la cuenta/i })).toBeVisible();
}
