import AxeBuilder from "@axe-core/playwright";
import { test as base, expect, type Page } from "@playwright/test";

export const LOCALES = [
	{ prefix: "", code: "en" },
	{ prefix: "/pt-br", code: "pt-BR" },
	{ prefix: "/es", code: "es" },
] as const;

type Fixtures = {
	/** Cookie-banner choice preset for the context; `null` leaves the banner up. */
	consent: "essential" | "all" | null;
	/** Browser errors collected during the test (CSP violations, uncaught exceptions). Checked after each test. */
	browserErrors: string[];
};

// Failing on these turns every spec into a CSP regression test: a blocked script/frame/connection shows up here.
const FATAL = [/content security policy/i, /refused to (load|execute|connect|frame|apply)/i];
// Network noise that is not an app bug (aborted navigations, BotID challenge off Vercel, analytics outside Vercel).
const IGNORED = [/net::ERR_ABORTED/, /_vercel\/(insights|speed-insights)/, /149e9513-01fa-4fb0-aad4-566afd725d1b/];

export const test = base.extend<Fixtures>({
	consent: ["essential", { option: true }],
	context: async ({ context, consent, baseURL }, use) => {
		if (consent && baseURL) {
			await context.addCookies([{ name: "consent", value: consent, url: baseURL, sameSite: "Lax" }]);
		}
		await use(context);
	},
	browserErrors: [
		async ({ page }, use) => {
			const errors: string[] = [];
			page.on("pageerror", (error) => errors.push(`pageerror: ${error.message}`));
			page.on("console", (message) => {
				const text = message.text();
				if (IGNORED.some((re) => re.test(text))) return;
				if (FATAL.some((re) => re.test(text))) errors.push(`console: ${text}`);
			});
			await use(errors);
			expect(errors, "browser errors (CSP violations / uncaught exceptions)").toEqual([]);
		},
		{ auto: true },
	],
});

export { expect };

/** axe-core WCAG 2.2 AA scan; fails on serious/critical violations and prints them readably. */
export async function expectAccessible(page: Page, options: { exclude?: string[] } = {}): Promise<void> {
	let builder = new AxeBuilder({ page }).withTags(["wcag2a", "wcag2aa", "wcag21a", "wcag21aa", "wcag22aa"]);
	for (const selector of options.exclude ?? []) builder = builder.exclude(selector);
	const { violations } = await builder.analyze();
	const blocking = violations.filter((v) => v.impact === "serious" || v.impact === "critical");
	const report = blocking.map(
		(v) => `${v.impact} ${v.id}: ${v.help}\n${v.nodes.map((n) => `  - ${n.target.join(" ")}`).join("\n")}`,
	);
	expect(report, `axe violations on ${page.url()}`).toEqual([]);
}
