import { createVerifiedAccount, expectLoggedIn, hasAuthSecret, logIn } from "../helpers/auth";
import { createItem, createMember, hasDatabase, sql } from "../helpers/db";
import { expect, expectAccessible, test } from "../helpers/fixtures";

test.describe("time to beat @auth", () => {
	test.skip(!hasAuthSecret() || !hasDatabase(), "needs E2E_BETTER_AUTH_SECRET and E2E_DATABASE_URL");

	test("report hours on a completed game; the median shows at the 3rd report", async ({ page, request, baseURL }) => {
		const item = await createItem();
		for (const hours of [10, 20]) {
			const member = await createMember("ttb");
			await sql()`insert into item_status (user_id, item_id, status) values (${member.id}, ${item.id}, 'completed')`;
			await sql()`insert into time_to_beat (user_id, item_id, main_minutes) values (${member.id}, ${item.id}, ${hours * 60})`;
		}
		const account = await createVerifiedAccount(request, baseURL ?? "", "ttb");
		await logIn(page, account);
		await expectLoggedIn(page);
		await page.goto(item.href);

		const block = page.getByRole("region", { name: "Time to beat" });
		await expect(block.getByText("2 reports")).toBeVisible();
		await expect(block.getByText(/Not enough reports yet/).first()).toBeVisible();

		const panel = page.getByRole("region", { name: "Your rating" });
		const form = panel.getByRole("form", { name: "Your time to beat" });
		await expect(form).toBeHidden();
		const status = panel.getByLabel("Status");
		await status.selectOption({ label: "Played" });
		await expect(form).toBeVisible();
		await form.getByLabel("Main story").fill("12");
		await expectAccessible(page);
		await form.getByRole("button", { name: "Save times" }).click();
		await expect(form.getByText("Times saved.")).toBeVisible();

		// Median of 10, 12, 20 = 12 h.
		await page.reload();
		await expect(block.getByText("3 reports")).toBeVisible();
		await expect(block.getByText("12 h")).toBeVisible();
		await expect(form.getByLabel("Main story")).toHaveValue("12");

		// Leaving "completed" deletes the report.
		await status.selectOption({ label: "Dropped" });
		await expect(form).toBeHidden();
		await expect
			.poll(
				async () =>
					(
						await sql()<
							{ n: number }[]
						>`select count(*)::int as n from time_to_beat where item_id = ${item.id}`
					)[0]?.n,
			)
			.toBe(2);
	});
});
