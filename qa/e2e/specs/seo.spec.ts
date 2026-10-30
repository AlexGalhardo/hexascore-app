import { itemLinks, openTopItem } from "../helpers/catalog";
import { expect, test } from "../helpers/fixtures";

test.describe("seo", () => {
	test("an item page carries its head tags and valid JSON-LD", async ({ page, prefix }) => {
		const href = await openTopItem(page, prefix, "games");
		const title = await page.title();
		expect(title.length).toBeGreaterThan(0);
		expect(title.length).toBeLessThanOrEqual(60);

		const description = await page.locator('meta[name="description"]').getAttribute("content");
		expect((description ?? "").length).toBeLessThanOrEqual(160);
		expect(
			await page.locator('meta[property="og:title"]').getAttribute("content"),
		).toBeTruthy();
		expect(
			await page.locator('meta[name="twitter:card"]').getAttribute("content"),
		).toBeTruthy();
		await expect(page.locator('meta[name="robots"]')).toHaveAttribute("content", /^index/);
		expect(await page.locator('link[rel="canonical"]').getAttribute("href")).toContain(
			href.replace(prefix, ""),
		);

		const raw = await page.locator('script[type="application/ld+json"]').first().textContent();
		const ld = JSON.parse(raw ?? "{}") as Record<string, unknown>;
		expect(ld["@context"]).toBe("https://schema.org");
		expect(ld["@type"]).toBe("VideoGame");
		expect(raw).not.toContain("</script");
		const rating = ld.aggregateRating as Record<string, unknown> | undefined;
		if (rating) {
			expect(rating.bestRating).toBe(10);
			expect(rating.worstRating).toBe(0.5);
			expect(Number(rating.ratingCount)).toBeGreaterThan(0);
		}
	});

	test("search is crawlable but noindex", async ({ page, prefix }) => {
		await page.goto(`${prefix}/search?q=a`);
		await expect(page.locator('meta[name="robots"]')).toHaveAttribute("content", /noindex/);
	});

	test("robots.txt blocks private areas and points at the sitemap", async ({ request }) => {
		const body = await (await request.get("/robots.txt")).text();
		for (const path of [
			"/admin",
			"/profile",
			"/me",
			"/billing",
			"/library",
			"/lists",
			"/login",
			"/signup",
		]) {
			expect(body).toContain(`Disallow: ${path}`);
		}
		expect(body).toContain("Disallow: /*?cursor=");
		expect(body).toMatch(/^Sitemap: https?:\/\/\S+\/sitemap\.xml$/m);
		expect(body).not.toMatch(/Disallow: \/search/);
	});

	test("the sitemap index lists children under 40,000 urls", async ({ request }) => {
		const index = await (await request.get("/sitemap.xml")).text();
		expect(index).toContain("<sitemapindex");
		const children = [...index.matchAll(/<loc>([^<]+)<\/loc>/g)].map(
			(m) => new URL(m[1]).pathname,
		);
		expect(children.length).toBeGreaterThan(1);
		for (const child of children) {
			const response = await request.get(child);
			expect(response.status(), child).toBe(200);
			const urls = ((await response.text()).match(/<url>/g) ?? []).length;
			expect(urls, child).toBeLessThanOrEqual(40_000);
		}
	});

	test("category pages are indexable and link their items", async ({ page, prefix }) => {
		await page.goto(`${prefix}/movies`);
		await expect(itemLinks(page, prefix, "movies").first()).toBeVisible();
		await expect(page.locator('meta[name="robots"]')).toHaveAttribute("content", /^index/);
	});
});
