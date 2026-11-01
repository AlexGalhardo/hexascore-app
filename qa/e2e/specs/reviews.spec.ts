import { createVerifiedAccount } from "../helpers/account";
import { expectNoSeriousA11yViolations } from "../helpers/axe";
import { openTopItem, projectSlot } from "../helpers/catalog";
import { expect, test } from "../helpers/fixtures";
import { canReadMail } from "../helpers/mail";

// Reviews need the social layer on: E2E_SOCIAL_OPEN=true with FLAG_SOCIAL=true on the app (SDD 14 §12.10).
const socialOn = process.env.E2E_SOCIAL_OPEN === "true";

test.describe("reviews", () => {
	test.skip(!canReadMail(), "needs docker access to the app container logs (E2E_APP_CONTAINER)");
	test.skip(!socialOn, "needs FLAG_SOCIAL=true on the app and E2E_SOCIAL_OPEN=true");

	test("write, edit, toggle the spoiler and delete a review", async ({
		page,
		prefix,
		t,
		any,
	}, testInfo) => {
		test.setTimeout(240_000);
		await createVerifiedAccount(page, prefix, any);
		// Its own slot offset keeps this spec off the items the ratings and profile specs rate in the same run.
		await openTopItem(page, prefix, "movies", projectSlot(testInfo) + 7);

		// Section ids sit on the heading (reviews-heading); the section is labelled by it.
		const reviews = page.locator("section[aria-labelledby=reviews-heading]");
		await expect(reviews.getByText(t("reviews.rateFirst"))).toBeVisible();

		const groups = page.getByRole("radiogroup");
		for (let i = 0; i < 6; i++) {
			await groups.nth(i).getByRole("radio").first().focus();
			await page.keyboard.press("End");
		}
		await page.getByRole("button", { name: t("ratings.save") }).click();
		await expect(page.getByRole("button", { name: any("ratings.update") })).toBeVisible();

		const body = reviews.getByLabel(t("reviews.body"));
		await body.fill("**Loved** it <script>window.pwned = 1</script>");
		await reviews.getByRole("button", { name: t("reviews.save") }).click();
		await expect(reviews.locator("strong", { hasText: "Loved" })).toBeVisible();
		await expect(
			reviews
				.locator("article")
				.getByText("<script>window.pwned = 1</script>", { exact: false }),
		).toBeVisible();
		expect(
			await page.evaluate(() => (window as unknown as { pwned?: number }).pwned),
		).toBeUndefined();
		await expectNoSeriousA11yViolations(page);

		await body.fill("Twist ending");
		await reviews.getByLabel(t("reviews.spoiler")).check();
		await reviews.getByRole("button", { name: t("reviews.save") }).click();
		await expect(reviews.getByRole("button", { name: t("reviews.reveal") })).toBeVisible();
		await expect(reviews.locator("article").getByText("Twist ending")).toHaveCount(0);
		await reviews.getByRole("button", { name: t("reviews.reveal") }).click();
		await expect(reviews.locator("article").getByText("Twist ending")).toBeVisible();

		await reviews.getByRole("button", { name: t("reviews.delete") }).click();
		await expect(reviews.getByText(t("reviews.empty"))).toBeVisible();
		await expectNoSeriousA11yViolations(page);
	});
});
