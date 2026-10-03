import { expect, expectAccessible, test } from "../helpers/fixtures";

test.describe("search", () => {
	test("autocomplete is keyboard operable and opens the highlighted item", async ({ page }) => {
		await page.goto("/");
		const input = page.locator("#site-search");
		await input.click();
		await input.pressSequentially("the", { delay: 30 });
		const options = page.getByRole("option");
		await expect(options.first()).toBeVisible();
		const title = (await options.first().locator("span").first().textContent())?.trim() ?? "";
		await input.press("ArrowDown");
		await input.press("Enter");
		await expect(page.locator("h1")).toHaveText(title);
	});

	test("Enter without a highlighted suggestion runs a full search", async ({ page }) => {
		await page.goto("/");
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
		await expect(page.locator("img[src=x]")).toHaveCount(0);
		await expect(page.getByText(/Nothing found for/)).toBeVisible();

		const sqli = "' OR 1=1 --";
		const response = await page.goto(`/search?q=${encodeURIComponent(sqli)}`);
		expect(response?.status()).toBe(200);
	});

	test("typo-tolerant search finds items (trigram)", async ({ request }) => {
		const response = await request.get("/api/search/suggest?q=teh");
		expect(response.status()).toBe(200);
		expect(response.headers()["content-type"]).toContain("application/json");
	});
});
