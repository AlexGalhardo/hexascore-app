import { defineConfig, devices } from "@playwright/test";

// Target: the built container (`docker compose --profile app up -d --wait`, seeded with DemoCatalog).
// The same suite runs against staging by pointing E2E_BASE_URL at it (D-21).
const baseURL = process.env.E2E_BASE_URL ?? "http://127.0.0.1:3000";

export const locales = ["en", "pt-BR", "es"] as const;

const devicesByProject = {
	desktop: { ...devices["Desktop Chrome"], viewport: { width: 1280, height: 800 } },
	mobile: { ...devices["Pixel 7"], viewport: { width: 360, height: 740 } },
};

export default defineConfig({
	testDir: "./specs",
	globalSetup: "./global-setup.ts",
	outputDir: "./test-results",
	// Signup, login and password reset solve a proof of work in the browser, so flows chain several of them.
	timeout: 120_000,
	expect: { timeout: 10_000 },
	fullyParallel: true,
	forbidOnly: Boolean(process.env.CI),
	retries: process.env.CI ? 1 : 0,
	workers: process.env.CI ? 2 : 3,
	reporter: process.env.CI ? [["list"], ["html", { open: "never" }], ["github"]] : [["list"], ["html", { open: "never" }]],
	use: {
		baseURL,
		// The production CSP carries upgrade-insecure-requests, which Chrome applies to the redirect hop of
		// an Inertia form post even on 127.0.0.1 (ERR_SSL_PROTOCOL_ERROR over plain HTTP). Specs that assert
		// the CSP itself opt back in with test.use({ bypassCSP: false }) and read the header instead.
		bypassCSP: true,
		trace: "retain-on-failure",
		screenshot: "only-on-failure",
	},
	projects: Object.entries(devicesByProject).flatMap(([device, use]) =>
		locales.map((locale) => ({ name: `${device}-${locale}`, use: { ...use, locale } })),
	),
});
