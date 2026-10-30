import type { Page } from "@playwright/test";
import { openTopItem } from "../helpers/catalog";
import { expect, test } from "../helpers/fixtures";

// The suite default bypasses the CSP (see playwright.config.ts); this spec runs with it enforced and
// fails on any violation the browser reports or any console error.
test.use({ bypassCSP: false });

async function watch(page: Page): Promise<string[]> {
	const problems: string[] = [];
	await page.addInitScript(() => {
		document.addEventListener("securitypolicyviolation", (e) => {
			console.error(`CSP violation: ${e.violatedDirective} ${e.blockedURI}`);
		});
	});
	page.on("console", (m) => {
		if (m.type() === "error") problems.push(m.text());
	});
	page.on("pageerror", (e) => problems.push(e.message));
	return problems;
}

test.describe("csp", () => {
	test("public pages run under the enforced policy with no violations", async ({
		page,
		prefix,
	}) => {
		const problems = await watch(page);
		for (const path of ["/", "/games", "/login", "/signup", "/docs", "/search?q=a"]) {
			await page.goto(`${prefix}${path}`);
			await page.waitForLoadState("networkidle");
		}
		await openTopItem(page, prefix, "games");
		await page.waitForLoadState("networkidle");
		expect(problems).toEqual([]);
	});

	test("the policy has a nonce, strict-dynamic and no unsafe-inline", async ({ request }) => {
		const csp = (await request.get("/")).headers()["content-security-policy"] ?? "";
		expect(csp).toMatch(/script-src [^;]*'nonce-/);
		expect(csp).toContain("'strict-dynamic'");
		expect(csp).not.toContain("'unsafe-inline'");
		expect(csp).not.toContain("'unsafe-eval'");
	});
});
