import { open } from "../helpers/account";
import { expectNoSeriousA11yViolations } from "../helpers/axe";
import { expect, test } from "../helpers/fixtures";

// Not ported: the localized /openapi*.json files (one English document, D-06).
const DOC_PAGES = [
	"authentication",
	"errors",
	"pagination",
	"rate-limits",
	"changes",
	"exports",
	"webhooks",
	"coverage",
	"changelog",
] as const;

test.describe("API docs", () => {
	test("the docs index and every guide page render", async ({ page, prefix }) => {
		await open(page, `${prefix}/docs`);
		await expect(page.locator("h1")).toBeVisible();
		await expectNoSeriousA11yViolations(page);

		for (const slug of DOC_PAGES) {
			const response = await page.goto(`${prefix}/docs/${slug}`);
			expect(response?.status(), slug).toBe(200);
			await expect(page.locator("h1")).toBeVisible();
		}
	});

	test("the Scalar reference is public and renders the OpenAPI document", async ({
		page,
		request,
	}) => {
		const doc = await request.get("/api/v1/openapi.json");
		expect(doc.status()).toBe(200);
		expect(((await doc.json()) as { openapi: string }).openapi).toMatch(/^3\.1\./);

		const response = await page.goto("/docs/reference");
		expect(response?.status()).toBe(200);
		await expect(page.getByText("HexaScore API").first()).toBeVisible({ timeout: 30_000 });
	});

	test("an unknown docs page is a 301 home", async ({ request, prefix }) => {
		expect((await request.get(`${prefix}/docs/nope`, { maxRedirects: 0 })).status()).toBe(301);
	});
});
