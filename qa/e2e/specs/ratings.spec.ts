import { createVerifiedAccount, reload } from "../helpers/account";
import { expectNoSeriousA11yViolations } from "../helpers/axe";
import { openTopItem, projectSlot } from "../helpers/catalog";
import { expect, test } from "../helpers/fixtures";
import { canReadMail } from "../helpers/mail";

// The legacy 0-100 spinbuttons became a half-star slider per category (D-00): the port drives it with the keyboard.
// A save redirects to an unprefixed URL (default locale), so everything after it matches copy in any locale.
test.describe("ratings", () => {
	test.skip(!canReadMail(), "needs docker access to the app container logs (E2E_APP_CONTAINER)");

	test("rate in six categories by keyboard, edit, then delete", async ({
		page,
		prefix,
		t,
		any,
	}, testInfo) => {
		await createVerifiedAccount(page, prefix, any);
		await openTopItem(page, prefix, "games", projectSlot(testInfo));

		const sliders = page.getByRole("slider");
		await expect(sliders).toHaveCount(6);
		await expect(page.getByRole("button", { name: t("ratings.save") })).toBeDisabled();

		for (let i = 0; i < 6; i++) {
			await sliders.nth(i).focus();
			await page.keyboard.press("End");
			await expect(sliders.nth(i)).toHaveAttribute("aria-valuenow", "10");
		}
		await expectNoSeriousA11yViolations(page);

		await page.getByRole("button", { name: t("ratings.save") }).click();
		const update = page.getByRole("button", { name: any("ratings.update") });
		await expect(update).toBeVisible();

		await reload(page);
		await expect(sliders.first()).toHaveAttribute("aria-valuenow", "10");

		await sliders.first().focus();
		await page.keyboard.press("ArrowLeft");
		await page.keyboard.press("ArrowLeft");
		await expect(sliders.first()).toHaveAttribute("aria-valuenow", "9");
		await update.click();
		await expect(update).toBeEnabled();

		await reload(page);
		await expect(sliders.first()).toHaveAttribute("aria-valuenow", "9");

		await page.getByRole("button", { name: any("ratings.delete") }).click();
		await expect(page.getByRole("button", { name: any("ratings.save") })).toBeVisible();
		await expect(sliders.first()).toHaveAttribute("aria-valuenow", "0");
	});

	test("item status can be set and cleared", async ({ page, prefix, any }, testInfo) => {
		await createVerifiedAccount(page, prefix, any);
		await openTopItem(page, prefix, "movies", projectSlot(testInfo));

		const status = page.locator("select:has(option[value=completed])");
		for (const value of ["want", "in_progress", "completed", "dropped"]) {
			await status.selectOption(value);
			await expect(status).toBeEnabled();
		}
		await reload(page);
		await expect(status).toHaveValue("dropped");

		await status.selectOption("");
		await expect(status).toBeEnabled();
		await reload(page);
		await expect(status).toHaveValue("");
	});
});
