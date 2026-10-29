import { expect, test } from "@playwright/test";
import { expectNoSeriousA11yViolations } from "../helpers/axe";
import { localePrefix } from "../helpers/locale";

test("home page renders in the project locale without serious a11y violations", async ({
	page,
}, testInfo) => {
	const locale = testInfo.project.use.locale ?? "en";
	const response = await page.goto(`${localePrefix(locale)}/`);
	expect(response?.status()).toBe(200);
	await expect(page.locator("html")).toHaveAttribute(
		"lang",
		new RegExp(`^${locale.split("-")[0]}`, "i"),
	);
	await expect(page.locator("main")).toBeVisible();
	await expectNoSeriousA11yViolations(page);
});
