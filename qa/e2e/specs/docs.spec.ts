import { expect, expectAccessible, test } from "../helpers/fixtures";

test.describe("API docs", () => {
	test("navigation reaches every docs page", async ({ page }) => {
		await page.goto("/docs");
		const nav = page.getByRole("navigation", { name: "Documentation" });
		for (const name of [
			"Authentication",
			"Errors",
			"Pagination",
			"Rate limits & quotas",
			"Changelog & versioning",
		]) {
			await nav.getByRole("link", { name }).click();
			await expect(page.locator("h1")).toBeVisible();
			await expect(nav.getByRole("link", { name })).toHaveAttribute("aria-current", "page");
		}
	});

	test("Scalar reference renders under the strict CSP", async ({ page }) => {
		await page.goto("/docs/reference");
		// Scalar renders each operation id/summary; "List games" style summaries come from openapi.json.
		await expect(page.getByText(/games/i).first()).toBeVisible({ timeout: 30_000 });
		await expect(page.getByText("Loading the API reference…")).toBeHidden();
		// Scalar ships its own markup; only our shell is held to axe here.
		await expectAccessible(page, { exclude: [".scalar-app", ".scalar-api-reference", "[class*='scalar']"] });
	});

	test("openapi.json is served as JSON", async ({ request }) => {
		const response = await request.get("/openapi.json");
		expect(response.status()).toBe(200);
		const spec = (await response.json()) as { openapi: string; paths: Record<string, unknown> };
		expect(spec.openapi).toMatch(/^3\./);
		expect(Object.keys(spec.paths).length).toBeGreaterThan(3);
	});
});
