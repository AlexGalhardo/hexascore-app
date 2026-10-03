import type { Page } from "@playwright/test";
import { createVerifiedAccount, expectLoggedIn, hasAuthSecret, logIn } from "../helpers/auth";
import { createItem, createMember, hasDatabase, insertRating, sql } from "../helpers/db";
import { expect, expectAccessible, test } from "../helpers/fixtures";

/** Saves the rating form and waits for the server to confirm (the form is optimistic). */
async function saveRating(page: Page, button: "Save rating" | "Update rating"): Promise<void> {
	const panel = page.getByRole("region", { name: "Your rating" });
	await panel.getByRole("button", { name: button }).click();
	await expect(panel.getByText("Rating saved.")).toBeVisible();
	await expect(panel.getByRole("button", { name: "Update rating" })).toBeDisabled();
}

test.describe("ratings @auth", () => {
	test.skip(!hasAuthSecret() || !hasDatabase(), "needs E2E_BETTER_AUTH_SECRET and E2E_DATABASE_URL");

	test("rate in 6 categories, edit, community score appears at the 5th rating", async ({
		page,
		request,
		baseURL,
	}) => {
		const item = await createItem();
		for (let i = 0; i < 4; i++) {
			const rater = await createMember("rater");
			await insertRating(rater.id, item.id, [60, 60, 60, 60, 60, 60]);
		}
		const account = await createVerifiedAccount(request, baseURL ?? "", "rate");
		await logIn(page, account);
		await expectLoggedIn(page);

		await page.goto(item.href);
		// The community score, rating count and threshold note live in the item header (cover · title · score).
		const community = page.getByRole("article").locator("header");
		await expect(community.getByText("4 ratings")).toBeVisible();
		await expect(community.getByText(/Not enough ratings/)).toBeVisible();
		await expect(community.getByRole("img", { name: "Not enough ratings yet" })).toBeVisible();

		const panel = page.getByRole("region", { name: "Your rating" });
		const fields = panel.getByRole("spinbutton");
		await expect(fields).toHaveCount(6);
		for (let i = 0; i < 6; i++) await fields.nth(i).fill("90");
		await expect(panel.getByRole("img", { name: "Score 90 out of 100" })).toBeVisible();
		await expectAccessible(page);
		await saveRating(page, "Save rating");

		// (4 × 60 + 90) / 5 = 66
		await page.reload();
		await expect(community.getByText("5 ratings")).toBeVisible();
		await expect(community.getByRole("img", { name: "Score 66 out of 100" })).toBeVisible();
		await expect(community.getByText(/Not enough ratings/)).toBeHidden();
		await expect(fields.first()).toHaveValue("90");

		// Edit: overall round((30 + 5 × 90) / 6) = 80 → (4 × 60 + 80) / 5 = 64
		await fields.first().fill("30");
		await expect(panel.getByRole("img", { name: "Score 80 out of 100" })).toBeVisible();
		await saveRating(page, "Update rating");
		await page.reload();
		await expect(community.getByText("5 ratings")).toBeVisible();
		await expect(community.getByRole("img", { name: "Score 64 out of 100" })).toBeVisible();
		await expect(fields.first()).toHaveValue("30");
	});

	test("item status cycles through the four values and shows on the profile shelves", async ({
		page,
		request,
		baseURL,
	}) => {
		const item = await createItem();
		const account = await createVerifiedAccount(request, baseURL ?? "", "status");
		await logIn(page, account);
		await expectLoggedIn(page);
		await page.goto(item.href);

		const status = page.getByRole("region", { name: "Your rating" }).getByLabel("Status");
		const stored = async () =>
			(
				await sql()<{ status: string }[]>`
					select s.status from item_status s join "user" u on u.id = s.user_id
					where u.email = ${account.email.toLowerCase()} and s.item_id = ${item.id}`
			)[0]?.status ?? null;
		for (const [label, value] of [
			["Want to play", "want"],
			["Playing", "in_progress"],
			["Played", "completed"],
			["Dropped", "dropped"],
		] as const) {
			await status.selectOption({ label });
			await expect(status).toBeEnabled();
			await expect.poll(stored).toBe(value);
		}
		await page.reload();
		await expect(status).toHaveValue("dropped");

		await page.goto(`/u/${account.nickname}?tab=status`);
		const shelf = page.getByRole("region", { name: /^Dropped/ });
		await expect(shelf.getByRole("link", { name: item.title })).toBeVisible();
		await expectAccessible(page);

		await page.goto(item.href);
		await status.selectOption({ label: "No status" });
		await expect.poll(stored).toBeNull();
	});

	test("curated list: create from the item page, then see it on the profile and the list page", async ({
		page,
		request,
		baseURL,
	}) => {
		const item = await createItem();
		const account = await createVerifiedAccount(request, baseURL ?? "", "lists");
		await logIn(page, account);
		await expectLoggedIn(page);
		await page.goto(item.href);

		const panel = page.getByRole("region", { name: "Your rating" });
		await panel.getByText("Add to list").click();
		await expect(panel.getByText("You have no game lists yet.")).toBeVisible();
		await expect(panel.getByText("Game lists: 0/10")).toBeVisible();
		const title = `E2E list ${item.id.slice(0, 8)}`;
		await panel.getByRole("textbox", { name: "Title" }).fill(title);
		await panel.getByRole("button", { name: "Create list" }).click();
		const entry = panel.getByRole("checkbox", { name: title });
		await expect(entry).toBeChecked();
		await expect(panel.getByText("Game lists: 1/10")).toBeVisible();
		await expectAccessible(page);

		await page.goto(`/u/${account.nickname}?tab=lists`);
		await expect(page.getByText(/1 item · Public/)).toBeVisible();
		// The new-list form requires a category and shows that category's quota.
		await expect(page.getByRole("combobox", { name: "Category" })).toHaveValue("game");
		await expect(page.getByText("Game lists: 1/10")).toBeVisible();
		await page.getByRole("link", { name: title }).click();
		await expect(page.getByRole("heading", { level: 1, name: title })).toBeVisible();
		await expect(page.getByRole("article").locator("header").getByText("Game", { exact: true })).toBeVisible();
		await expect(page.getByRole("link", { name: item.title })).toBeVisible();
		await expectAccessible(page);

		// Unchecking removes the item from the list.
		await page.goto(item.href);
		await panel.getByText("Add to list").click();
		await entry.uncheck();
		await expect(entry).toBeEnabled();
		await expect
			.poll(
				async () =>
					(
						await sql()<
							{ n: number }[]
						>`select count(*)::int as n from list_items where item_id = ${item.id}`
					)[0]?.n,
			)
			.toBe(0);
	});
});
