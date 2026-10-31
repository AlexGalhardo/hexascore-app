import { expect, test } from "../helpers/fixtures";

test.describe("i18n", () => {
	test("the page language, hreflang set and language switcher follow the URL prefix", async ({
		page,
		prefix,
		locale,
		t,
	}) => {
		await page.goto(`${prefix}/login`);
		await expect(page.locator("html")).toHaveAttribute("lang", locale);
		await expect(page.getByRole("heading", { level: 1 })).toHaveText(t("auth.login.title"));

		await page.goto(`${prefix}/games`);
		const hreflangs = await page
			.locator('link[rel="alternate"][hreflang]')
			.evaluateAll((els) => els.map((el) => el.getAttribute("hreflang")).sort());
		expect(hreflangs).toEqual(["en", "es", "pt-BR", "x-default"]);

		await page.goto(`${prefix}/login`);
		// The switcher lives in the header menu, closed by default (its links are not in the a11y tree).
		await page.getByLabel(t("nav.more"), { exact: true }).click();
		const switcher = page.getByRole("list", { name: t("nav.language") });
		await expect(switcher.getByRole("link")).toHaveCount(3);
		await switcher.getByRole("link", { name: "ES" }).click();
		await expect(page).toHaveURL(/\/es\/login$/);
		await expect(page.locator("html")).toHaveAttribute("lang", "es");
	});

	test("a locale cookie redirects an unprefixed page to its prefix", async ({
		playwright,
		baseURL,
	}) => {
		const api = await playwright.request.newContext({
			baseURL,
			extraHTTPHeaders: { Cookie: "locale=pt-BR" },
		});
		const response = await api.get("/games", { maxRedirects: 0 });
		expect(response.status()).toBe(307);
		expect(response.headers().location).toBe("/pt-br/games");
		await api.dispose();
	});

	test("an unknown prefix is a 404", async ({ request }) => {
		expect((await request.get("/xx/games")).status()).toBe(404);
	});

	test("the catalogue renders translated copy, not message keys", async ({ page, prefix }) => {
		await page.goto(`${prefix}/games`);
		// Text nodes outside script/style: a leaked message key is a node on its own.
		const nodes = await page.locator("body").evaluate((body) => {
			const out: string[] = [];
			const walker = document.createTreeWalker(body, NodeFilter.SHOW_TEXT);
			for (let n = walker.nextNode(); n; n = walker.nextNode()) {
				if (!n.parentElement?.closest("script,style,noscript"))
					out.push(n.textContent ?? "");
			}
			return out;
		});
		expect(nodes.filter((line) => /^[a-z]+(\.[a-zA-Z]+)+$/.test(line.trim()))).toEqual([]);
	});
});
