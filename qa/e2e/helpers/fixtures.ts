import { test as base } from "@playwright/test";
import { anyLocale, translator } from "./t";
import { localePrefix } from "./locale";

export { expect } from "@playwright/test";

// A distinct client IP per test (honoured because the e2e container trusts the Docker gateway as proxy),
// so the per-IP login and signup limits of one spec never throttle another.
function fakeIp(): string {
	const octet = (): number => 1 + Math.floor(Math.random() * 254);
	return `10.${octet()}.${octet()}.${octet()}`;
}

export const test = base.extend<{ t: ReturnType<typeof translator>; any: typeof anyLocale; prefix: string; locale: string }>({
	extraHTTPHeaders: async ({}, use) => use({ "X-Forwarded-For": fakeIp() }),
	locale: async ({}, use, testInfo) => use(String(testInfo.project.use.locale ?? "en")),
	t: async ({ locale }, use) => use(translator(locale)),
	any: async ({}, use) => use(anyLocale),
	prefix: async ({ locale }, use) => use(localePrefix(locale)),
});
