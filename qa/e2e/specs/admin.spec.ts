import { createVerifiedAccount } from "../helpers/account";
import { signIn } from "../helpers/auth";
import { expect, test } from "../helpers/fixtures";
import { canReadMail } from "../helpers/mail";

// Not ported: "a member reports a user; an admin resolves it" needs the report-user dialog, which the vukem
// profile page does not have yet; the admin side is covered by the Go tests (tests/admin_*).
test.describe("admin", () => {
	test("/admin sends visitors to the login form", async ({ page }) => {
		for (const path of ["/admin", "/admin/reports", "/admin/users"]) {
			await page.goto(path);
			await expect(page, path).toHaveURL(/\/login/);
		}
	});

	test("/admin answers 404 to regular members", async ({ page, prefix, any }) => {
		test.skip(
			!canReadMail(),
			"needs docker access to the app container logs (E2E_APP_CONTAINER)",
		);
		await createVerifiedAccount(page, prefix, any);
		for (const path of ["/admin", "/admin/reports", "/admin/users"]) {
			expect((await page.goto(path))?.status(), path).toBe(404);
		}
	});

	test("the seeded admin reaches the panel and its reports list", async ({ page, prefix }) => {
		await signIn(page, prefix);
		expect((await page.goto("/admin"))?.status()).toBe(200);
		const reports = await page.goto("/admin/reports");
		expect(reports?.status()).toBe(200);
		await expect(page.locator("h1")).toBeVisible();
	});
});
