import { expect, type Locator, type Page, type TestInfo } from "@playwright/test";
import { open } from "./account";

export const SECTIONS = ["games", "movies", "series", "books"] as const;
export type Section = (typeof SECTIONS)[number];

// Catalogue entries of one section: the item cards link to `/<section>/<slug>`.
export function itemLinks(page: Page, prefix: string, section: Section): Locator {
	return page.locator(`main ul a[href^="${prefix}/${section}/"]`);
}

// Opens the nth most rated item of a section; the DemoCatalog seeds those with enough ratings for a community score.
// Specs that write ratings pass a per-project nth so parallel runs never trip the per-item burst limit (HTTP 429).
export async function openTopItem(
	page: Page,
	prefix: string,
	section: Section,
	nth = 0,
): Promise<string> {
	await open(page, `${prefix}/${section}?sort=-rating_count`);
	const link = itemLinks(page, prefix, section).nth(nth);
	await expect(link).toBeVisible();
	const href = (await link.getAttribute("href")) ?? "";
	await open(page, href);
	await expect(page.locator("h1")).toBeVisible();
	return href;
}

// One slot per Playwright project, so each project rates a different item.
export function projectSlot(testInfo: TestInfo): number {
	return Math.max(
		0,
		testInfo.config.projects.findIndex((p) => p.name === testInfo.project.name),
	);
}
