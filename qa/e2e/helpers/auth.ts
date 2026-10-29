import type { Page } from "@playwright/test";
import { DEMO_ADMIN_EMAIL, DEMO_PASSWORD } from "./db";

// Signs in through the real login form so the session is created by the app itself.
export async function signIn(page: Page, prefix: string, email = DEMO_ADMIN_EMAIL): Promise<void> {
	await page.goto(`${prefix}/login`);
	await page.getByLabel(/e-?mail/i).fill(email);
	await page.locator('input[type="password"]').first().fill(DEMO_PASSWORD);
	await page.locator("form button[type=submit]").first().click();
	await page.waitForURL((url) => !url.pathname.endsWith("/login"));
}
