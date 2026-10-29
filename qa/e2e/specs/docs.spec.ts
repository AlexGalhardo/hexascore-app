import { open } from "../helpers/account";
import { expectNoSeriousA11yViolations } from "../helpers/axe";
import { expect, test } from "../helpers/fixtures";

// Not ported: the Scalar reference page and the localized /openapi*.json files have no route in the vukem
// app yet (the specs are exported to openapi/ and checked by scripts/openapi-check.sh instead).
const DOC_PAGES = ["authentication", "errors", "pagination", "rate-limits", "changelog"] as const;

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

	test("an unknown docs page is a 404", async ({ page, prefix }) => {
		expect((await page.goto(`${prefix}/docs/nope`))?.status()).toBe(404);
	});
});
