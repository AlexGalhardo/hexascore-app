import { expect, test } from "../helpers/fixtures";

// Generous on purpose: the suite runs on a loaded Windows host with 3 workers. Lighthouse owns the strict
// LCP 2500 ms and 170 KB script budgets (Q-A20); this only catches gross regressions in server or render time.
const TTFB_BUDGET_MS = Number(process.env.E2E_BUDGET_TTFB_MS ?? 800);
const LOAD_BUDGET_MS = Number(process.env.E2E_BUDGET_LOAD_MS ?? 4000);

const PAGES = ["/", "/games", "/docs", "/pricing"];

test.describe("performance budget", () => {
	for (const path of PAGES) {
		test(`${path} meets the TTFB and load budget`, async ({ page, prefix }) => {
			await page.goto(`${prefix}${path}`, { waitUntil: "load" });
			const timing = await page.evaluate(() => {
				const [nav] = performance.getEntriesByType(
					"navigation",
				) as PerformanceNavigationTiming[];
				return {
					ttfb: nav.responseStart - nav.startTime,
					load: nav.loadEventEnd - nav.startTime,
				};
			});
			expect(timing.ttfb, "time to first byte (ms)").toBeLessThan(TTFB_BUDGET_MS);
			expect(timing.load, "load event (ms)").toBeLessThan(LOAD_BUDGET_MS);
		});
	}
});
