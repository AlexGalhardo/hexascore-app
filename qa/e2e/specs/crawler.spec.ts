import type { Page } from "@playwright/test";
import { expect, test } from "../helpers/fixtures";

const MAX_PAGES = 60;

// Authenticated areas redirect to /login, auth pages are under redesign, and the rest are not HTML pages
// (JSON, images, redirects to third parties). Their links are still checked, just not crawled.
const NOT_CRAWLED =
	/^\/(login|signup|2fa|logout|go|api|covers|og|sitemaps?|auth|challenge|turnstile|csp-reports|u|suggest|suggestions|lists|profile|me|admin|billing|user|two-factor-challenge)(\/|$)/;

function normalise(url: URL): string {
	url.hash = "";
	url.search = url.search === "?" ? "" : url.search;
	return url.pathname.replace(/\/$/, "") + url.search || "/";
}

async function hrefs(page: Page): Promise<string[]> {
	return page.$$eval("a[href]", (anchors) => anchors.map((a) => a.getAttribute("href") ?? ""));
}

test.describe("crawler", () => {
	test("every public page reachable from home answers without errors and every internal link resolves", async ({
		page,
		request,
		prefix,
	}) => {
		test.setTimeout(300_000);

		let pageErrors: string[] = [];
		page.on("console", (message) => {
			if (message.type() === "error") pageErrors.push(`${message.text()}`);
		});
		page.on("pageerror", (error) => pageErrors.push(error.message));

		const home = await page.goto(`${prefix}/`);
		expect(home?.status()).toBe(200);
		const origin = new URL(page.url()).origin;

		const queue = [`${prefix}/`];
		const crawled = new Set<string>();
		const linkChecked = new Set<string>();
		const broken: string[] = [];
		const failures: string[] = [];

		while (queue.length > 0 && crawled.size < MAX_PAGES) {
			const path = queue.shift() as string;
			const key = normalise(new URL(path, origin));
			if (crawled.has(key)) continue;
			crawled.add(key);

			pageErrors = [];
			const response = await page.goto(path, { waitUntil: "load" });
			const status = response?.status() ?? 0;
			if (status >= 400) {
				broken.push(`${path} -> ${status}`);
				continue;
			}
			if (pageErrors.length > 0) {
				failures.push(`${path}: ${pageErrors.join(" | ")}`);
			}

			for (const href of await hrefs(page)) {
				if (!href || href.startsWith("mailto:") || href.startsWith("tel:")) continue;
				const target = new URL(href, page.url());
				if (target.origin !== origin) continue;
				const targetKey = normalise(target);
				if (linkChecked.has(targetKey)) {
					if (!crawled.has(targetKey) && !NOT_CRAWLED.test(target.pathname)) {
						queue.push(`${target.pathname}${target.search}`);
					}
					continue;
				}
				linkChecked.add(targetKey);

				const check = await request.get(`${target.pathname}${target.search}`);
				if (check.status() >= 400) {
					broken.push(`${path} links to ${targetKey} -> ${check.status()}`);
				}
				if (!NOT_CRAWLED.test(target.pathname)) {
					queue.push(`${target.pathname}${target.search}`);
				}
			}
		}

		expect(crawled.size, "crawl reached more than the home page").toBeGreaterThan(5);
		expect(broken, "broken pages or internal links").toEqual([]);
		expect(failures, "console errors on crawled pages").toEqual([]);
	});
});
