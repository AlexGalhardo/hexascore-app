import { defineConfig, devices } from "@playwright/test";

// Target: a running deployment. Locally `next start` (see docs/testing.md → E2E); in CI the Vercel preview of the PR.
const baseURL = process.env.E2E_BASE_URL ?? "http://localhost:3000";
const bypass = process.env.VERCEL_AUTOMATION_BYPASS_SECRET;

const desktop = { ...devices["Desktop Chrome"], viewport: { width: 1280, height: 800 } };
const mobile = {
	...devices["Pixel 7"],
	viewport: { width: 360, height: 740 },
	screen: { width: 360, height: 740 },
	deviceScaleFactor: 2,
};

// Account flows hit Better Auth's per-IP rate limit (3 sign-ins / 10 s); they run in desktop-light only,
// so a full run does not trip it.
const AUTH = /@auth/;

export default defineConfig({
	testDir: "./specs",
	globalSetup: "./global-setup.ts",
	outputDir: "./test-results",
	timeout: 60_000,
	expect: { timeout: 10_000 },
	fullyParallel: true,
	forbidOnly: Boolean(process.env.CI),
	retries: process.env.CI ? 1 : 0,
	workers: process.env.CI ? 2 : 3,
	reporter: process.env.CI
		? [["list"], ["html", { open: "never" }], ["github"]]
		: [["list"], ["html", { open: "never" }]],
	use: {
		baseURL,
		locale: "en-US",
		trace: "retain-on-failure",
		screenshot: "only-on-failure",
		extraHTTPHeaders: bypass
			? { "x-vercel-protection-bypass": bypass, "x-vercel-set-bypass-cookie": "samesitenone" }
			: {},
	},
	projects: [
		{ name: "desktop-light", use: { ...desktop, colorScheme: "light" } },
		{ name: "desktop-dark", use: { ...desktop, colorScheme: "dark" }, grepInvert: AUTH },
		{ name: "mobile-light", use: { ...mobile, colorScheme: "light" }, grepInvert: AUTH },
		{ name: "mobile-dark", use: { ...mobile, colorScheme: "dark" }, grepInvert: AUTH },
	],
});
