import { open } from "../helpers/account";
import { expectNoSeriousA11yViolations } from "../helpers/axe";
import { expect, test } from "../helpers/fixtures";

test.describe("search", () => {
	test("autocomplete is keyboard operable and opens the highlighted item", async ({
		page,
		prefix,
		t,
	}) => {
		await open(page, `${prefix}/`);
		const input = page.getByRole("combobox", { name: t("nav.search") });
		await input.click();
		await input.pressSequentially("labore", { delay: 30 });
		const options = page.getByRole("option");
		await expect(options.first()).toBeVisible();
		await input.press("ArrowDown");
		await input.press("Enter");
		await expect(page).toHaveURL(
			new RegExp(`${prefix}/(games|movies|series|books|people|companies|franchises)/`),
		);
		await expect(page.locator("h1")).toBeVisible();
	});

	test("Enter without a highlighted suggestion runs a full search", async ({
		page,
		prefix,
		t,
	}) => {
		await open(page, `${prefix}/`);
		const input = page.getByRole("combobox", { name: t("nav.search") });
		await input.fill("labore");
		await input.press("Enter");
		await expect(page).toHaveURL(
			(url) => url.pathname + url.search === `${prefix}/search?q=labore`,
		);
		await expect(page.getByRole("heading", { level: 1 })).toBeVisible();
		await expectNoSeriousA11yViolations(page);
	});

	test("short, empty and hostile queries are handled", async ({ page, prefix, t }) => {
		await open(page, `${prefix}/search?q=a`);
		await expect(page.getByText(t("search.tooShort"))).toBeVisible();

		const payload = `"><img src=x onerror=alert(1)>`;
		await page.goto(`${prefix}/search?q=${encodeURIComponent(payload)}`);
		await expect(page.locator("h1")).toContainText(payload);
		await expect(page.locator("img[src=x]")).toHaveCount(0);

		for (const q of ["' OR 1=1 --", "\u0000\u0000", "😀😀😀"]) {
			const response = await page.goto(`${prefix}/search?q=${encodeURIComponent(q)}`);
			expect(response?.status(), q).toBe(200);
		}

		// An over-long query is refused with a client error, never a server error.
		const long = await page.goto(`${prefix}/search?q=${"a".repeat(500)}`);
		expect(long?.status()).toBe(400);
	});

	test("typo-tolerant suggestions answer JSON", async ({ request }) => {
		const response = await request.get("/search/suggest?q=teh");
		expect(response.status()).toBe(200);
		expect(response.headers()["content-type"]).toContain("application/json");
	});
});
