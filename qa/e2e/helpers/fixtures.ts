import { test as base } from "@playwright/test";
import { localePrefix } from "./locale";
import { anyLocale, translator } from "./t";

export { expect } from "@playwright/test";

// A distinct client IP per test (honoured because the e2e container trusts the Docker gateway as proxy),
// so the per-IP login and signup limits of one spec never throttle another.
function fakeIp(): string {
	const octet = (): number => 1 + Math.floor(Math.random() * 254);
	return `10.${octet()}.${octet()}.${octet()}`;
}

export const test = base.extend<{
	t: ReturnType<typeof translator>;
	any: typeof anyLocale;
	prefix: string;
	locale: string;
}>({
	// biome-ignore lint/correctness/noEmptyPattern: Playwright requires destructuring the first fixture argument
	extraHTTPHeaders: async ({}, use) => use({ "X-Forwarded-For": fakeIp() }),
	// biome-ignore lint/correctness/noEmptyPattern: Playwright requires destructuring the first fixture argument
	locale: async ({}, use, testInfo) => use(String(testInfo.project.use.locale ?? "en")),
	t: async ({ locale }, use) => use(translator(locale)),
	// biome-ignore lint/correctness/noEmptyPattern: Playwright requires destructuring the first fixture argument
	any: async ({}, use) => use(anyLocale),
	prefix: async ({ locale }, use) => use(localePrefix(locale)),
});
