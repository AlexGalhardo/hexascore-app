import { expect, test } from "bun:test";
import { readFileSync } from "node:fs";

const root = new URL("../../", import.meta.url);

// The QA app runs with APP_ENV=production, so CanonicalHost 301s every URL whose Host is not the
// APP_URL host, and Lighthouse then measures a redirect before each page (LCP over budget in CI).
test("the Lighthouse default base URL is the compose APP_URL origin", async () => {
	delete process.env.E2E_BASE_URL;
	const rc = (await import(new URL("lighthouserc.cjs", root).href)).default as { ci: { collect: { url: string[] } } };
	const compose = readFileSync(new URL("docker-compose.yml", root), "utf8");
	const appURL = compose.match(/APP_URL: (http:\/\/[^$\s]+):\$\{QA_PORT:-(\d+)\}/);
	expect(appURL).not.toBeNull();
	const origin = `${appURL?.[1]}:${appURL?.[2]}`;
	for (const url of rc.ci.collect.url) {
		expect(new URL(url).origin).toBe(origin);
	}
});
