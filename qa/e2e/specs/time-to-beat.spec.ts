import { createVerifiedAccount, reload } from "../helpers/account";
import { expectNoSeriousA11yViolations } from "../helpers/axe";
import { openTopItem, projectSlot } from "../helpers/catalog";
import { expect, test } from "../helpers/fixtures";
import { canReadMail } from "../helpers/mail";

// Only a player who marked the game completed can report hours; a save redirects to an unprefixed URL.
test.describe("time to beat", () => {
	test.skip(!canReadMail(), "needs docker access to the app container logs (E2E_APP_CONTAINER)");

	test("report hours on a completed game, then lose the form when the status changes", async ({
		page,
		prefix,
		t,
		any,
	}, testInfo) => {
		await createVerifiedAccount(page, prefix, any);
		await openTopItem(page, prefix, "games", projectSlot(testInfo));

		const block = page.locator("section[aria-labelledby=ttb-heading]");
		await expect(block).toBeVisible();
		const main = block.getByLabel(t("item.timeToBeat.styles.main"));
		await expect(main).toHaveCount(0);

		const status = page.locator("select:has(option[value=completed])");
		await status.selectOption("completed");
		await expect(main).toBeVisible();
		await main.fill("12");
		await expectNoSeriousA11yViolations(page);
		const saved = page.waitForResponse(
			(r) => r.url().endsWith("/time-to-beat") && r.request().method() === "PUT",
		);
		await block.getByRole("button", { name: t("item.timeToBeat.save") }).click();
		expect((await saved).status()).toBeLessThan(400);

		await reload(page);
		await expect(block.getByLabel(any("item.timeToBeat.styles.main"))).toHaveValue("12");

		await page.locator("select:has(option[value=completed])").selectOption("dropped");
		await expect(block.getByLabel(any("item.timeToBeat.styles.main"))).toHaveCount(0);
	});
});
