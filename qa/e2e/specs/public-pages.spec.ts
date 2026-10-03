import { expect, expectAccessible, LOCALES, test } from "../helpers/fixtures";
import { openNavMenu } from "../helpers/session";

const PAGES = [
	"/",
	"/games",
	"/movies",
	"/series",
	"/books",
	"/search?q=the",
	"/pricing",
	"/contact",
	"/login",
	"/signup",
	"/recover-password",
	"/policy",
	"/terms",
	"/docs",
	"/docs/authentication",
	"/docs/errors",
	"/docs/pagination",
	"/docs/rate-limits",
	"/docs/changelog",
];

for (const { prefix, code } of LOCALES) {
	test.describe(`public pages (${code})`, () => {
		for (const path of PAGES) {
			test(`${prefix}${path} renders, is themed and accessible`, async ({ page }, info) => {
				const response = await page.goto(`${prefix}${path}`);
				expect(response?.status()).toBe(200);
				await expect(page.locator("html")).toHaveAttribute("lang", code);
				const scheme = info.project.use.colorScheme === "dark" ? "dark" : "light";
				await expect(page.locator("html")).toHaveAttribute("data-theme", scheme);
				await expect(page.locator("h1").first()).toBeVisible();
				await expectAccessible(page);
			});
		}
	});
}

test("theme toggle persists across reloads", async ({ page }, info) => {
	await page.goto("/");
	const initial = info.project.use.colorScheme === "dark" ? "dark" : "light";
	const flipped = initial === "dark" ? "light" : "dark";
	await openNavMenu(page);
	await page.getByRole("button", { name: "Toggle theme" }).click();
	await expect(page.locator("html")).toHaveAttribute("data-theme", flipped);
	await page.reload();
	await expect(page.locator("html")).toHaveAttribute("data-theme", flipped);
});

test.describe("cookie banner", () => {
	test.use({ consent: null });

	test("asks once and remembers the choice", async ({ page, context }) => {
		await page.goto("/");
		const banner = page.getByRole("region", { name: "Cookies" });
		await expect(banner).toBeVisible();
		await expectAccessible(page);
		await banner.getByRole("button", { name: "Essential only" }).click();
		await expect(banner).toBeHidden();
		const cookies = await context.cookies();
		expect(cookies.find((c) => c.name === "consent")?.value).toBe("essential");
		await page.reload();
		await expect(page.getByRole("region", { name: "Cookies" })).toBeHidden();
	});
});

test.describe("language switch", () => {
	test("locale prefixes render translated navigation", async ({ page }) => {
		await page.goto("/pt-br");
		await expect(page.getByRole("link", { name: "Jogos" }).first()).toBeVisible();
		await page.goto("/es");
		await expect(page.getByRole("link", { name: "Juegos" }).first()).toBeVisible();
	});
});
