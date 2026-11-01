import type { Page } from "@playwright/test";
import { expect, test } from "../helpers/fixtures";

// Crawlers and no-JS clients get the server-rendered page: content, one <h1>, links, no client rendering.
test.use({ javaScriptEnabled: false });

async function firstHref(page: Page, pattern: RegExp): Promise<string> {
	const hrefs = await page
		.locator("a[href]")
		.evaluateAll((els) => els.map((a) => a.getAttribute("href") ?? ""));
	const found = hrefs.find((h) => pattern.test(h));
	expect(found, `no link matching ${pattern}`).toBeTruthy();
	return found as string;
}

test.describe("ssr without javascript", () => {
	for (const path of [
		"/",
		"/games",
		"/movies",
		"/series",
		"/books",
		"/docs",
		"/terms",
		"/privacy",
		"/contact",
		"/login",
		"/signup",
	]) {
		test(`${path} renders its content`, async ({ page, prefix }) => {
			const response = await page.goto(`${prefix}${path}`);
			expect(response?.status()).toBe(200);
			await expect(page.locator("h1")).toHaveCount(1);
			expect(
				((await page.locator("main").textContent()) ?? "").trim().length,
			).toBeGreaterThan(20);
		});
	}

	test("an item page of every type shows its title and structured data", async ({
		page,
		prefix,
	}) => {
		for (const section of ["games", "movies", "series", "books"]) {
			await page.goto(`${prefix}/${section}?sort=-rating_count`);
			const href = await firstHref(page, new RegExp(`^${prefix}/${section}/[^/?#]+$`));
			const response = await page.goto(href);
			expect(response?.status(), href).toBe(200);
			await expect(page.locator("h1")).toHaveCount(1);
			expect(
				((await page.locator("h1").textContent()) ?? "").trim().length,
				href,
			).toBeGreaterThan(0);
			await expect(page.locator('script[type="application/ld+json"]')).toHaveCount(1);
		}
	});

	test("an entity page lists its items", async ({ page, prefix }) => {
		await page.goto(`${prefix}/games?sort=-rating_count`);
		await page.goto(await firstHref(page, new RegExp(`^${prefix}/games/[^/?#]+$`)));
		const entity = await firstHref(
			page,
			new RegExp(`^${prefix}/(people|companies|franchises|platforms|genres)/[^/?#]+$`),
		);
		const response = await page.goto(entity);
		expect(response?.status(), entity).toBe(200);
		await expect(page.locator("h1")).toHaveCount(1);
		expect(await page.locator("main a[href]").count()).toBeGreaterThan(0);
	});

	test("search results render server-side", async ({ page, prefix }) => {
		const response = await page.goto(`${prefix}/search?q=a`);
		expect(response?.status()).toBe(200);
		await expect(page.locator("h1")).toHaveCount(1);
	});
});
