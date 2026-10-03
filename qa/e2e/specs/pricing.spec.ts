import { createVerifiedAccount, hasAuthSecret, logIn } from "../helpers/auth";
import { expect, expectAccessible, test } from "../helpers/fixtures";

test.describe("pricing", () => {
	test("lists the three plans and switches currency on the server", async ({ page }) => {
		await page.goto("/pricing");
		for (const plan of ["Free", "Pro", "Business"]) {
			await expect(page.getByRole("heading", { name: plan, exact: true })).toBeVisible();
		}
		const currency = page.getByRole("navigation", { name: "Currency" });
		await currency.getByRole("link", { name: "EUR" }).click();
		await expect(page).toHaveURL(/currency=EUR/);
		await expect(page.getByText("€").first()).toBeVisible();
		await currency.getByRole("link", { name: "BRL" }).click();
		await expect(page.getByText("R$").first()).toBeVisible();
		await expectAccessible(page);
	});

	test("a forged currency falls back to a valid one", async ({ page }) => {
		const response = await page.goto("/pricing?currency=XXX");
		expect(response?.status()).toBe(200);
		await expect(page.getByRole("heading", { name: "Pro", exact: true })).toBeVisible();
	});

	test("checkout requires login", async ({ page }) => {
		await page.goto("/profile/api/checkout?plan=pro");
		await expect(page).toHaveURL(/\/login\?next=/);
	});
});

test.describe("checkout page states @auth", () => {
	test.skip(!hasAuthSecret(), "needs E2E_BETTER_AUTH_SECRET to create a verified account");

	test("shows the current plan, the provider form or the not-configured notice, and return states", async ({
		page,
		request,
		baseURL,
	}) => {
		const account = await createVerifiedAccount(request, baseURL ?? "", "checkout");
		await logIn(page, account);
		await page.waitForURL((url) => !url.pathname.endsWith("/login"));

		await page.goto("/profile/api/checkout?plan=pro&currency=USD");
		await expect(page.getByRole("heading", { name: "Plan and billing" })).toBeVisible();
		await expect(page.getByText("Current plan")).toBeVisible();
		const configured = await page.getByRole("button", { name: /Continue to payment|Generate PIX code/ }).count();
		if (!configured) await expect(page.getByText("Payments are not configured on this server.")).toBeVisible();
		await expectAccessible(page);

		await page.goto("/profile/api/checkout?status=canceled");
		await expect(page.getByText("Checkout canceled. Nothing was charged.")).toBeVisible();
		await page.goto("/profile/api/checkout?status=success");
		await expect(page.getByText(/Payment received/)).toBeVisible();
	});
});
