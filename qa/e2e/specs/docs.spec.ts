import { expect, expectAccessible, test } from "../helpers/fixtures";

test.describe("API docs", () => {
	test("navigation reaches every docs page", async ({ page }) => {
		await page.goto("/docs");
		const nav = page.getByRole("navigation", { name: "Documentation" });
		await nav.getByRole("link", { name: "Authentication" }).click();
		await expect(nav.getByRole("link", { name: "Authentication" })).toHaveAttribute("aria-current", "page");
		// The nav scrolls horizontally on mobile; follow the remaining hrefs instead of chasing off-screen links.
		for (const name of ["Errors", "Pagination", "Rate limits & quotas", "Changelog & versioning"]) {
			const href = await nav.getByRole("link", { name }).getAttribute("href");
			await page.goto(href ?? "/docs");
			await expect(page.locator("h1")).toBeVisible();
			await expect(nav.getByRole("link", { name })).toHaveAttribute("aria-current", "page");
		}
	});

	test("Scalar reference renders under the strict CSP", async ({ page }) => {
		await page.goto("/docs/reference");
		// Scalar renders each operation id/summary; "List games" style summaries come from openapi.json.
		// Scalar collapses its sidebar on mobile, so match the first rendered occurrence, not the first in the DOM.
		await expect(page.getByRole("main").getByText(/games/i).filter({ visible: true }).first()).toBeVisible({
			timeout: 30_000,
		});
		await expect(page.getByText("Loading the API reference…")).toBeHidden();
		// Scalar ships its own markup; only our shell is held to axe here.
		await expectAccessible(page, { exclude: [".scalar-app", ".scalar-api-reference", "[class*='scalar']"] });
	});

	test("localized reference loads the translated spec", async ({ page }) => {
		await page.goto("/pt-br/docs/reference");
		await expect(page.getByRole("main").getByText("Listar jogos").filter({ visible: true }).first()).toBeVisible({
			timeout: 30_000,
		});
	});

	for (const [path, heading, section] of [
		["/pt-br/docs", "Primeiros passos", "1. Crie uma chave de API"],
		["/es/docs", "Primeros pasos", "1. Crea una clave de API"],
	] as const) {
		test(`${path} guide is translated`, async ({ page }) => {
			await page.goto(path);
			await expect(page.locator("h1")).toHaveText(heading);
			await expect(page.getByRole("heading", { level: 2, name: section })).toBeVisible();
			await expectAccessible(page);
		});
	}

	for (const path of ["/openapi.json", "/openapi.pt-BR.json", "/openapi.es.json"]) {
		test(`${path} is served as JSON`, async ({ request }) => {
			const response = await request.get(path);
			expect(response.status()).toBe(200);
			const spec = (await response.json()) as { openapi: string; paths: Record<string, unknown> };
			expect(spec.openapi).toMatch(/^3\./);
			expect(Object.keys(spec.paths).length).toBeGreaterThan(3);
		});
	}
});
