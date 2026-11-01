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

		const groups = page.getByRole("radiogroup");
		await expect(groups).toHaveCount(6);
		const checked = (i: number) => groups.nth(i).locator("input:checked");
		await expect(page.getByRole("button", { name: t("ratings.save") })).toBeDisabled();

		for (let i = 0; i < 6; i++) {
			await groups.nth(i).getByRole("radio").first().focus();
			await page.keyboard.press("End");
			await expect(checked(i)).toHaveValue("10");
		}
		await expectNoSeriousA11yViolations(page);

		await page.getByRole("button", { name: t("ratings.save") }).click();
		const update = page.getByRole("button", { name: any("ratings.update") });
		await expect(update).toBeVisible();

		await reload(page);
		await expect(checked(0)).toHaveValue("10");

		await checked(0).focus();
		await page.keyboard.press("ArrowLeft");
		await page.keyboard.press("ArrowLeft");
		await expect(checked(0)).toHaveValue("9");
		await update.click();
		await expect(update).toBeEnabled();

		await reload(page);
		await expect(checked(0)).toHaveValue("9");

		await page.getByRole("button", { name: any("ratings.delete") }).click();
		await expect(page.getByRole("button", { name: any("ratings.save") })).toBeVisible();
		await expect(checked(0)).toHaveCount(0);
	});

	test("item status can be set and cleared", async ({ page, prefix, t, any }, testInfo) => {
		await createVerifiedAccount(page, prefix, any);
		await openTopItem(page, prefix, "movies", projectSlot(testInfo));

		const status = page.getByRole("group", { name: t("social.status.label") });
		for (const value of ["want", "in_progress", "completed", "dropped"]) {
			const button = status.getByRole("button", { name: t(`social.status.movie.${value}`) });
			await button.click();
			await expect(button).toHaveAttribute("aria-pressed", "true");
		}
		const dropped = status.getByRole("button", { name: t("social.status.movie.dropped") });
		await reload(page);
		await expect(dropped).toHaveAttribute("aria-pressed", "true");

		await dropped.click();
		await expect(dropped).toHaveAttribute("aria-pressed", "false");
		await reload(page);
		await expect(status.locator("[aria-pressed=true]")).toHaveCount(0);
	});
});
