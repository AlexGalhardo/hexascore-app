import { expectNoSeriousA11yViolations } from "../helpers/axe";
import { openTopItem, SECTIONS } from "../helpers/catalog";
import { expect, test } from "../helpers/fixtures";

// Not ported: the "users' scores" modal and trailer consent of the legacy item page have no counterpart in
// the vukem app yet (SDD 02 §10 lists them for a later phase); critic scores are dropped (D-09).
test.describe("item page", () => {
	for (const section of SECTIONS) {
		test(`${section}: shows the community score with category averages and attribution`, async ({
			page,
			prefix,
			locale,
			t,
		}) => {
			await openTopItem(page, prefix, section);
			await expect(page.locator("html")).toHaveAttribute(
				"lang",
				new RegExp(`^${locale.split("-")[0]}`, "i"),
			);
			await expect(page.locator("h1")).toBeVisible();
			await expect(
				page.getByRole("heading", { name: t("item.communityScore") }),
			).toBeVisible();
			await expect(
				page.getByRole("heading", { name: t("item.categoryScores") }),
			).toBeAttached();
			await expect(page.getByText(t("item.attribution"))).toBeVisible();
			await expectNoSeriousA11yViolations(page);
		});
	}

	test("a guest is asked to log in to rate", async ({ page, prefix, t }) => {
		await openTopItem(page, prefix, "games");
		await expect(page.getByText(t("item.loginToRate"))).toBeVisible();
		await expect(page.getByRole("slider")).toHaveCount(0);
	});

	test("an unknown slug is a 404 page", async ({ page, prefix }) => {
		const response = await page.goto(`${prefix}/games/does-not-exist-0000`);
		expect(response?.status()).toBe(404);
		await expect(page.locator("h1")).toBeVisible();
	});
});
