import { itemLinks, SECTIONS } from "../helpers/catalog";
import { expect, expectAccessible, test } from "../helpers/fixtures";

test.describe("catalog", () => {
	for (const section of SECTIONS) {
		test(`${section}: filters change the URL and the list`, async ({ page }) => {
			await page.goto(`/${section}`);
			const filters = page.getByRole("form", { name: "Filters" });
			await expect(itemLinks(page, section).first()).toBeVisible();
			const before = await itemLinks(page, section).first().getAttribute("href");

			await filters.getByLabel("Sort by").selectOption("score_desc");
			await filters.getByLabel("Min. score").fill("60");
			await filters.getByRole("button", { name: "Apply" }).click();
			await expect(page).toHaveURL(/sort=score_desc/);
			await expect(page).toHaveURL(/min_score=60/);
			await expect(itemLinks(page, section).first()).toBeVisible();
			expect(await itemLinks(page, section).first().getAttribute("href")).not.toBe(before);
			await expectAccessible(page);

			await filters.getByRole("link", { name: "Reset" }).click();
			await expect(page).not.toHaveURL(/min_score/);
		});
	}

	test("keyset pagination goes forward and back", async ({ page }) => {
		await page.goto("/movies?sort=popularity_desc");
		const firstPage = await itemLinks(page, "movies").first().getAttribute("href");
		const pages = page.getByRole("navigation", { name: "Pages" });
		await pages.getByRole("link", { name: /Next/ }).click();
		await expect(page).toHaveURL(/cursor=/);
		await expect(itemLinks(page, "movies").first()).not.toHaveAttribute("href", firstPage ?? "");
		await pages.getByRole("link", { name: /Previous/ }).click();
		await expect(itemLinks(page, "movies").first()).toHaveAttribute("href", firstPage ?? "");
	});

	test("filters with no match show the empty state", async ({ page }) => {
		await page.goto("/books?year=1800");
		await expect(page.getByText("No items match these filters.")).toBeVisible();
	});

	test("a tampered cursor or filter never breaks the page", async ({ page }) => {
		for (const query of ["cursor=%27%20OR%201%3D1--", "min_score=abc", "sort=<script>", "year=-1", "genre=%00"]) {
			const response = await page.goto(`/games?${query}`);
			expect(response?.status(), query).toBeLessThan(500);
		}
	});
});
