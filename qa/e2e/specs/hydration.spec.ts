import { expect, test } from "../helpers/fixtures";

test.describe("hydration feedback", () => {
	for (const path of ["/login", "/signup"]) {
		test(`${path} submit button shows a pending state while the page is inert`, async ({
			page,
			prefix,
		}) => {
			await page.goto(`${prefix}${path}`);
			const spinner = page.locator("[data-submit] .submit-spinner");
			await expect(spinner).toBeHidden();

			// The loader keeps <html inert> until Inertia has hydrated; a click then is swallowed.
			await page.evaluate(() => {
				document.documentElement.inert = true;
			});
			await expect(spinner).toBeVisible();

			await page.evaluate(() => {
				document.documentElement.inert = false;
			});
			await expect(spinner).toBeHidden();
		});
	}

	test("login form posts to the login URL, never as a GET", async ({ page, prefix }) => {
		await page.goto(`${prefix}/login`);
		const form = page.locator("form[method='post']");
		await expect(form).toHaveAttribute("action", new RegExp("/login$"));
	});
});
