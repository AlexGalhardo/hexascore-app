import type { Page } from "@playwright/test";
import { open } from "./account";
import { DEMO_ADMIN_EMAIL, DEMO_PASSWORD } from "./db";

// Signs in through the real login form so the session is created by the app itself.
// open() waits for React props: a click before hydration reloads /login and never leaves it.
export async function signIn(page: Page, prefix: string, email = DEMO_ADMIN_EMAIL): Promise<void> {
	await open(page, `${prefix}/login`);
	await page
		.locator("main form input:not([type=password]):not([type=hidden]):not([type=checkbox])")
		.first()
		.fill(email);
	await page.locator('input[type="password"]').first().fill(DEMO_PASSWORD);
	await page.locator("form button[type=submit]").first().click();
	await page.waitForURL((url) => !url.pathname.endsWith("/login"));
}
