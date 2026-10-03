import { expect, type Page } from "@playwright/test";

/** Mobile hides the navbar actions behind the hamburger; opens it when it is shown, no-op on desktop. */
export async function openNavMenu(page: Page): Promise<void> {
	const menu = page.locator('button[aria-controls="nav-menu"]');
	if ((await menu.isVisible()) && (await menu.getAttribute("aria-expanded")) !== "true") await menu.click();
}

export async function logOut(page: Page): Promise<void> {
	await openNavMenu(page);
	await page.getByRole("button", { name: "Account menu" }).click();
	await page.getByRole("menuitem", { name: "Log out" }).click();
	// On mobile the navbar link sits inside the closed hamburger menu: attached, not visible.
	await expect(page.getByRole("link", { name: "Log in" }).first()).toBeAttached();
}

/** Enables TOTP 2FA from /profile and returns the secret shown next to the QR code. */
export async function enableTwoFactor(
	page: Page,
	password: string,
	code: (secret: string) => Promise<string>,
): Promise<string> {
	await page.goto("/profile");
	const section = page
		.locator("section")
		.filter({ has: page.getByRole("heading", { name: "Two-factor authentication" }) });
	await section.getByRole("button", { name: "Enable 2FA" }).click();
	await section.getByLabel("Confirm with your password").fill(password);
	await section.getByRole("button", { name: "Enable 2FA" }).click();
	const secret = ((await section.locator("code").first().textContent()) ?? "").trim();
	expect(secret).toMatch(/^[A-Z2-7]+=*$/);
	await section.getByLabel("6-digit code").fill(await code(secret));
	await section.getByRole("button", { name: "Verify and enable" }).click();
	await expect(section.getByRole("heading", { name: "Backup codes" })).toBeVisible();
	await section.getByRole("button", { name: "I saved my codes" }).click();
	await expect(section.getByText("Two-factor authentication is on.")).toBeVisible();
	return secret;
}
