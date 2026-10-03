import { createVerifiedAccount, expectLoggedIn, hasAuthSecret, logIn } from "../helpers/auth";
import { createMember, grantAdmin, hasDatabase, sql } from "../helpers/db";
import { expect, expectAccessible, test } from "../helpers/fixtures";
import { logOut } from "../helpers/session";

// Non-admins get the regular 404 (app/[locale]/admin/guard.ts) so the area is not revealed.
test.describe("admin @auth", () => {
	test.skip(!hasAuthSecret() || !hasDatabase(), "needs E2E_BETTER_AUTH_SECRET and E2E_DATABASE_URL");

	test("/admin answers 404 to visitors and regular members", async ({ page, request, baseURL }) => {
		for (const path of ["/admin", "/admin/reports"]) {
			expect((await page.goto(path))?.status(), path).toBe(404);
		}
		const account = await createVerifiedAccount(request, baseURL ?? "", "noadmin");
		await logIn(page, account);
		await expectLoggedIn(page);
		for (const path of ["/admin", "/admin/reports"]) {
			expect((await page.goto(path))?.status(), path).toBe(404);
			await expect(page.getByRole("navigation", { name: "Admin" })).toBeHidden();
		}
	});

	test("a member reports a user; an admin resolves it in /admin/reports", async ({ page, request, baseURL }) => {
		const target = await createMember("reported");
		const reporter = await createVerifiedAccount(request, baseURL ?? "", "reporter");
		const admin = await createVerifiedAccount(request, baseURL ?? "", "admin");
		await grantAdmin(admin.email);

		await logIn(page, reporter);
		await expectLoggedIn(page);
		await page.goto(`/u/${target.nickname}`);
		await page.getByRole("button", { name: "Report user" }).click();
		const dialog = page.getByRole("dialog", { name: "Report" });
		await dialog.getByLabel("Reason").selectOption({ label: "Impersonation" });
		await dialog.getByLabel("Details (optional)").fill("E2E report");
		await expectAccessible(page);
		await dialog.getByRole("button", { name: "Send report" }).click();
		await expect(dialog.getByText("Thanks. A moderator will review it.")).toBeVisible();
		await dialog.getByRole("button", { name: "Close" }).click();
		await logOut(page);

		await logIn(page, admin);
		await expectLoggedIn(page);
		const response = await page.goto("/admin/reports");
		expect(response?.status()).toBe(200);
		await expect(page.getByRole("heading", { level: 1, name: "Open reports" })).toBeVisible();
		const row = page.getByRole("row").filter({ hasText: target.nickname });
		await expect(row).toContainText("impersonation");
		await expect(row).toContainText(reporter.nickname);
		await expectAccessible(page);
		await row.getByRole("button", { name: "Resolve" }).click();
		await expect(row.getByRole("status")).toHaveText("done");

		const status = async () =>
			(
				await sql()<{ status: string }[]>`
					select status from reports where target_type = 'user' and target_id = ${target.id}`
			)[0]?.status;
		await expect.poll(status).toBe("resolved");
		await page.goto("/admin/reports?status=closed");
		await expect(page.getByRole("row").filter({ hasText: target.nickname })).toContainText("resolved");
	});
});
