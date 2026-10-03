import type { Locator, Page } from "@playwright/test";

export const SECTIONS = ["games", "movies", "series", "books"] as const;

/** Visible item links of a catalog page (desktop table or mobile list, whichever is rendered). */
export function itemLinks(page: Page, section: string, prefix = ""): Locator {
	return page.locator(`main a[href^="${prefix}/${section}/"]:visible`);
}

/** Opens the most popular item of a section (always seeded with ratings and, for most types, trailers). */
export async function openTopItem(page: Page, section = "games", prefix = ""): Promise<string> {
	await page.goto(`${prefix}/${section}?sort=popularity_desc`);
	const link = itemLinks(page, section, prefix).first();
	const href = (await link.getAttribute("href")) ?? "";
	await link.click();
	await page.waitForURL(`**${href}`);
	return href;
}
