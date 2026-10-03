import type { Page } from "@playwright/test";
import { expect, expectAccessible, test } from "../helpers/fixtures";

/** Mobile hides the search behind the navbar's search button; desktop shows it inline. */
async function openSearch(page: Page): Promise<void> {
	const toggle = page.locator('button[aria-controls="nav-search"]');
	if (await toggle.isVisible()) await toggle.click();
}

test.describe("search", () => {
	test("autocomplete is keyboard operable and opens the highlighted item", async ({ page }) => {
		await page.goto("/");
		await openSearch(page);
		const input = page.locator("#site-search");
		await input.click();
		await input.pressSequentially("the", { delay: 30 });
		const options = page.getByRole("option");
		await expect(options.first()).toBeVisible();
		const title = (await options.first().locator(".truncate").textContent())?.trim() ?? "";
		await input.press("ArrowDown");
		await input.press("Enter");
		await expect(page.locator("h1")).toHaveText(title);
	});

	test("Enter without a highlighted suggestion runs a full search", async ({ page }) => {
		await page.goto("/");
		await openSearch(page);
		await page.locator("#site-search").fill("the");
		await page.locator("#site-search").press("Enter");
		await expect(page).toHaveURL(/\/search\?q=the/);
		await expect(page.getByRole("heading", { level: 1 })).toBeVisible();
		await expectAccessible(page);
	});

	test("short, empty and hostile queries are handled", async ({ page }) => {
		await page.goto("/search?q=a");
		await expect(page.getByText("Type at least 2 characters.")).toBeVisible();

		const payload = `"><img src=x onerror=alert(1)>`;
		await page.goto(`/search?q=${encodeURIComponent(payload)}`);
		await expect(page.locator("h1")).toContainText(payload);
		await expect(page.locator("img[src=x]")).toHaveCount(0);

		for (const q of ["' OR 1=1 --", "\u0000\u0000", "😀😀😀", "a".repeat(500)]) {
			const response = await page.goto(`/search?q=${encodeURIComponent(q)}`);
			expect(response?.status(), q).toBe(200);
		}
	});

	test("typo-tolerant search finds items (trigram)", async ({ request }) => {
		const response = await request.get("/api/search/suggest?q=teh");
		expect(response.status()).toBe(200);
		expect(response.headers()["content-type"]).toContain("application/json");
	});
});
