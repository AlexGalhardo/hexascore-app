import { expect, test } from "../helpers/fixtures";

test.describe("keyboard", () => {
	test("the skip link is the first tab stop and moves focus to main", async ({
		page,
		prefix,
		t,
	}) => {
		await page.goto(`${prefix}/games`);
		await page.waitForFunction(() => document.activeElement?.tagName === "H1");
		// The heading holds focus after load; reset the sequential focus start to the document top.
		await page.evaluate(() => {
			document.body.tabIndex = -1;
			document.body.focus();
			document.body.removeAttribute("tabindex");
		});
		await page.keyboard.press("Tab");
		await expect(page.getByRole("link", { name: t("nav.skip") })).toBeFocused();
		await page.keyboard.press("Enter");
		await expect(page).toHaveURL(/#main$/);
		await expect(page.locator("main")).toBeVisible();
	});

	test("header controls are reachable in order and show a focus indicator", async ({
		page,
		prefix,
	}) => {
		await page.goto(`${prefix}/`);
		// The loader keeps the document inert until hydration, so Tab hits nothing before React.
		await page.waitForFunction(() =>
			Object.keys(document.querySelector("header a") ?? {}).some((key) =>
				key.startsWith("__reactProps$"),
			),
		);
		const seen = new Set<string>();
		for (let i = 0; i < 25; i++) {
			await page.keyboard.press("Tab");
			const info = await page.evaluate(() => {
				const el = document.activeElement as HTMLElement | null;
				if (!el || el === document.body) return null;
				const style = getComputedStyle(el);
				return {
					tag: el.tagName,
					text: (el.getAttribute("aria-label") ?? el.textContent ?? "")
						.trim()
						.slice(0, 30),
					ring:
						(style.outlineStyle !== "none" && style.outlineWidth !== "0px") ||
						style.boxShadow !== "none",
				};
			});
			if (!info) continue;
			expect(info.ring, `no focus indicator on ${info.tag} "${info.text}"`).toBe(true);
			seen.add(`${info.tag}:${info.text}`);
		}
		expect(seen.size).toBeGreaterThan(5);
	});

	test("opening a page with the keyboard moves focus to its heading", async ({
		page,
		prefix,
	}) => {
		await page.goto(`${prefix}/`);
		await page.waitForFunction(() =>
			Object.keys(document.querySelector("header a") ?? {}).some((k) =>
				k.startsWith("__reactProps$"),
			),
		);
		const link = page.locator(`header a[href^="${prefix}/games"]`).first();
		await link.focus();
		await page.keyboard.press("Enter");
		await expect(page).toHaveURL(/\/games$/);
		await expect(page.locator("h1")).toBeFocused();
	});
});
