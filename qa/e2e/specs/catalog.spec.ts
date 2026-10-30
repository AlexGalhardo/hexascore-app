import { open } from "../helpers/account";
import { expectNoSeriousA11yViolations } from "../helpers/axe";
import { itemLinks, openTopItem, SECTIONS } from "../helpers/catalog";
import { expect, test } from "../helpers/fixtures";

test.describe("catalog", () => {
	for (const section of SECTIONS) {
		test(`${section}: filters change the URL and the list`, async ({ page, prefix, t }) => {
			await open(page, `${prefix}/${section}`);
			await expect(itemLinks(page, prefix, section).first()).toBeVisible();
			await expectNoSeriousA11yViolations(page);

			const form = page.locator("main form[method=get]").first();
			await form.getByLabel(t("catalog.sort")).selectOption({ value: "-score" });
			await form.getByLabel(t("catalog.minScore")).fill("1");
			await form.getByRole("button", { name: t("catalog.apply") }).click();
			await expect(page).toHaveURL(/sort=-score/);
			await expect(page).toHaveURL(/min_score=1/);
			// A 1-star floor hides items whose stored score is under 10. Either result is a real list.
			await expect(
				itemLinks(page, prefix, section)
					.first()
					.or(page.getByText(t("catalog.empty"))),
			).toBeVisible();

			await page.getByRole("link", { name: t("catalog.reset") }).click();
			await expect(page).not.toHaveURL(/min_score/);
		});
	}

	test("filters with no match show the empty state", async ({ page, prefix, t }) => {
		await open(page, `${prefix}/books?year=1800`);
		await expect(page.getByText(t("catalog.empty"))).toBeVisible();
		await expectNoSeriousA11yViolations(page);
	});

	test("a tampered cursor or filter never breaks the page", async ({ page, prefix }) => {
		for (const query of [
			"cursor=%27%20OR%201%3D1--",
			"min_score=abc",
			"sort=<script>",
			"year=-1",
			"genre=%00",
		]) {
			const response = await page.goto(`${prefix}/games?${query}`);
			expect(response?.status(), query).toBeLessThan(500);
		}
	});

	test("an item opens from the list with its title and an accessible page", async ({
		page,
		prefix,
		locale,
	}) => {
		await openTopItem(page, prefix, "movies");
		await expect(page.locator("html")).toHaveAttribute(
			"lang",
			new RegExp(`^${locale.split("-")[0]}`, "i"),
		);
		await expectNoSeriousA11yViolations(page);
	});
});
