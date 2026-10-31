import { createVerifiedAccount, open, signOut } from "../helpers/account";
import { expectNoSeriousA11yViolations } from "../helpers/axe";
import { expect, test } from "../helpers/fixtures";
import { canReadMail } from "../helpers/mail";

test.describe("passkeys", () => {
	// WebAuthn refuses an IP address as RP ID; the QA app's APP_URL host is localhost.
	test.use({ baseURL: process.env.E2E_BASE_URL ?? "http://localhost:3000" });
	test.skip(!canReadMail(), "needs docker access to the app container logs (E2E_APP_CONTAINER)");

	test("add a passkey and sign in with it", async ({ page, prefix, t }) => {
		const account = await createVerifiedAccount(page, prefix, t);
		const client = await page.context().newCDPSession(page);
		await client.send("WebAuthn.enable");
		await client.send("WebAuthn.addVirtualAuthenticator", {
			options: {
				protocol: "ctap2",
				transport: "internal",
				hasResidentKey: true,
				hasUserVerification: true,
				isUserVerified: true,
				automaticPresenceSimulation: true,
			},
		});

		await open(page, `${prefix}/profile`);
		await expectNoSeriousA11yViolations(page);

		const section = page.locator("#passkeys");
		await section
			.getByLabel(t("account.passkeys.password"), { exact: true })
			.fill(account.password);
		await section.getByRole("button", { name: t("account.passkeys.add") }).click();
		await expect(
			section.getByRole("button", {
				name: t("account.passkeys.removeLabel", { name: "Passkey" }),
			}),
		).toBeVisible();

		await signOut(page);
		await open(page, `${prefix}/login`);
		await page.getByRole("button", { name: t("auth.passkey") }).click();
		await expect(page).toHaveURL(/\/dashboard/);

		await open(page, `${prefix}/profile`);
		const remove = page
			.locator("#passkeys")
			.getByRole("button", { name: t("account.passkeys.removeLabel", { name: "Passkey" }) });
		await remove.click();
		await expect(remove).toHaveCount(0);
		await expect(
			page.locator("#passkeys").getByText(t("account.passkeys.empty")),
		).toBeVisible();
	});
});
