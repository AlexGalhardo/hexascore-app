import { open } from "../helpers/account";
import { expectNoSeriousA11yViolations } from "../helpers/axe";
import { expect, test } from "../helpers/fixtures";

// /pricing stays a public page; live Stripe keys still wait on Q-H31.
const PAGES = [
	"/",
	"/games",
	"/movies",
	"/series",
	"/books",
	"/search?q=labore",
	"/contact",
	"/login",
	"/signup",
	"/recover-password",
	"/privacy",
	"/terms",
	"/docs",
	"/docs/authentication",
	"/docs/errors",
	"/docs/pagination",
	"/docs/rate-limits",
	"/docs/changelog",
];

test.describe("public pages", () => {
	for (const path of PAGES) {
		test(`${path} renders in the project locale and is accessible`, async ({
			page,
			prefix,
			locale,
		}) => {
			const response = await page.goto(`${prefix}${path}`);
			expect(response?.status()).toBe(200);
			await expect(page.locator("html")).toHaveAttribute(
				"lang",
				new RegExp(`^${locale.split("-")[0]}`, "i"),
			);
			await expect(page.locator("h1").first()).toBeVisible();
			await expectNoSeriousA11yViolations(page);
		});
	}

	test("the locale prefixes render translated navigation", async ({ page, prefix, t }) => {
		await open(page, `${prefix}/`);
		await expect(
			page.getByRole("link", { name: t("nav.games"), exact: true }).first(),
		).toBeVisible();
	});

	test("theme toggle persists across reloads", async ({ page, prefix, t }) => {
		await open(page, `${prefix}/`);
		const html = page.locator("html");
		const before = await html.evaluate((el) => el.classList.contains("dark"));
		await page.getByRole("button", { name: t("nav.theme") }).click();
		await expect.poll(() => html.evaluate((el) => el.classList.contains("dark"))).toBe(!before);
		await page.reload();
		await expect.poll(() => html.evaluate((el) => el.classList.contains("dark"))).toBe(!before);
	});
});

test.describe("cookie banner", () => {
	test.use({ consent: null });

	test("asks once and remembers the choice", async ({ page, context, prefix, t }) => {
		await open(page, `${prefix}/`);
		const banner = page.getByRole("region", { name: "Cookies" });
		await expect(banner).toBeVisible();
		await expectNoSeriousA11yViolations(page);
		await banner.getByRole("button", { name: t("cookies.decline") }).click();
		await expect(banner).toBeHidden();
		const cookies = await context.cookies();
		expect(cookies.find((c) => c.name === "consent")?.value).toBe("essential");
		await page.reload();
		await expect(page.getByRole("region", { name: "Cookies" })).toBeHidden();
	});
});
