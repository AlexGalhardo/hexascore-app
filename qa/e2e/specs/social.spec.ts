import { createVerifiedAccount, signOut } from "../helpers/account";
import { expectNoSeriousA11yViolations } from "../helpers/axe";
import { openTopItem, projectSlot } from "../helpers/catalog";
import { expect, test } from "../helpers/fixtures";
import { canReadMail } from "../helpers/mail";

// Not ported (D-06): following a member and /feed. Report-a-user has no UI yet (see admin.spec.ts).
test.describe("profiles", () => {
	test.skip(!canReadMail(), "needs docker access to the app container logs (E2E_APP_CONTAINER)");

	test("a public profile lists ratings; a private one hides them from everyone but the owner", async ({
		page,
		prefix,
		t,
		any,
	}, testInfo) => {
		const account = await createVerifiedAccount(page, prefix, any);
		// Slot offset keeps this spec off the items the ratings spec rates in the same run.
		const itemHref = await openTopItem(page, prefix, "books", projectSlot(testInfo));
		const title = ((await page.locator("h1").textContent()) ?? "").trim();

		const groups = page.getByRole("radiogroup");
		for (let i = 0; i < 6; i++) {
			await groups.nth(i).getByRole("radio").first().focus();
			await page.keyboard.press("End");
		}
		await page.getByRole("button", { name: t("ratings.save") }).click();
		await expect(page.getByRole("button", { name: any("ratings.update") })).toBeVisible();

		const profile = `${prefix}/u/${account.nickname.toLowerCase()}`;
		await page.goto(profile);
		await expect(page.getByRole("heading", { level: 1, name: account.nickname })).toBeVisible();
		await expect(page.getByRole("link", { name: title }).first()).toBeVisible();
		await expectNoSeriousA11yViolations(page);

		await page.goto(`${prefix}/profile`);
		const privacy = page.locator("#privacy");
		await privacy.getByLabel(t("account.privacy.label")).check();
		const saved = page.waitForResponse(
			(r) => r.url().endsWith("/profile/privacy") && r.request().method() === "PATCH",
		);
		await privacy.locator("button[type=submit]").click();
		expect((await saved).status()).toBeLessThan(400);

		await page.goto(profile);
		await expect(page.getByText(t("social.profile.privateOwner"))).toBeVisible();
		await expect(page.getByRole("link", { name: title }).first()).toBeVisible();

		await signOut(page);
		await page.goto(profile);
		await expect(page.getByText(t("social.profile.private"))).toBeVisible();
		await expect(page.getByRole("link", { name: title })).toHaveCount(0);
		await expectNoSeriousA11yViolations(page);
		expect(itemHref).toContain(`${prefix}/books/`);
	});
});
