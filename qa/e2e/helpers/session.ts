import { expect, type Page } from "@playwright/test";

export async function logOut(page: Page): Promise<void> {
	await page.getByRole("button", { name: "Account menu" }).click();
	await page.getByRole("menuitem", { name: "Log out" }).click();
	await expect(page.getByRole("link", { name: "Log in" }).first()).toBeVisible();
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
