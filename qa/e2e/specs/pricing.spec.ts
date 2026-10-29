import { expectNoSeriousA11yViolations } from "../helpers/axe";
import { expect, test } from "../helpers/fixtures";

test.describe("pricing", () => {
	test("lists the three plans, the coverage statement and switches currency on the server", async ({
		page,
		prefix,
		t,
	}) => {
		await page.goto(`${prefix}/pricing`);
		for (const plan of ["free", "pro", "business"]) {
			await expect(
				page.getByRole("heading", { name: t(`billing.plans.${plan}`), exact: true }),
			).toBeVisible();
		}
		await expect(
			page.getByRole("heading", { name: t("billing.pricing.coverageTitle") }),
		).toBeVisible();

		const currency = page.getByRole("navigation", { name: t("billing.pricing.currency") });
		await currency.getByRole("link", { name: "EUR" }).click();
		await expect(page).toHaveURL(/currency=eur/);
		await expect(page.getByText("€").first()).toBeVisible();
		await currency.getByRole("link", { name: "BRL" }).click();
		await expect(page.getByText(/(R\$|BRL)\s?29/).first()).toBeVisible();
		await expectNoSeriousA11yViolations(page);
	});

	test("a forged currency falls back to a valid one", async ({ page, prefix, t }) => {
		const response = await page.goto(`${prefix}/pricing?currency=XXX`);
		expect(response?.status()).toBe(200);
		await expect(
			page.getByRole("heading", { name: t("billing.plans.pro"), exact: true }),
		).toBeVisible();
	});

	test("a guest is sent to sign up for a paid plan", async ({ page, prefix, t }) => {
		await page.goto(`${prefix}/pricing`);
		await page
			.getByRole("link", { name: t("billing.pricing.ctaSignup") })
			.first()
			.click();
		await expect(page).toHaveURL(/\/signup/);
	});
});
