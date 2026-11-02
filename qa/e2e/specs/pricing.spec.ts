import { expectNoSeriousA11yViolations } from "../helpers/axe";
import { expect, test } from "../helpers/fixtures";
import { escapeRegExp } from "../helpers/t";

test.describe("pricing", () => {
	test("lists the three plans, the coverage link and switches currency on the server", async ({
		page,
		prefix,
		t,
	}) => {
		await page.goto(`${prefix}/pricing`);
		for (const plan of ["free", "hobby", "pro"]) {
			await expect(
				page.getByRole("heading", {
					name: new RegExp(`^${escapeRegExp(String(t(`billing.plans.${plan}`)))}\\b`),
				}),
			).toBeVisible();
		}
		await expect(
			page.getByRole("link", { name: t("billing.pricing.coverageLink") }),
		).toHaveAttribute("href", `${prefix}/docs/coverage`);

		const currency = page.getByRole("navigation", { name: t("billing.pricing.currency") });
		await currency.getByRole("link", { name: "EUR" }).click();
		await expect(page).toHaveURL(/currency=eur/);
		await expect(page.getByText("€").first()).toBeVisible();
		await currency.getByRole("link", { name: "BRL" }).click();
		await expect(page.getByText(/R\$\s?50/).first()).toBeVisible();
		await expectNoSeriousA11yViolations(page);
	});

	test("a forged currency falls back to a valid one", async ({ page, prefix, t }) => {
		const response = await page.goto(`${prefix}/pricing?currency=XXX`);
		expect(response?.status()).toBe(200);
		await expect(
			page.getByRole("heading", {
				name: new RegExp(`^${escapeRegExp(String(t("billing.plans.pro")))}\\b`),
			}),
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
