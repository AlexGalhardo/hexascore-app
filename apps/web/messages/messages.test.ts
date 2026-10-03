import { describe, expect, test } from "bun:test";
import { NAMESPACES } from "@/i18n/request";
import { routing } from "@/i18n/routing";
import { loadLegal } from "@/lib/legal";

function keys(value: unknown, prefix = ""): string[] {
	if (typeof value !== "object" || value === null || Array.isArray(value)) return [prefix];
	return Object.entries(value).flatMap(([k, v]) => keys(v, prefix ? `${prefix}.${k}` : k));
}

async function load(locale: string, ns: string): Promise<unknown> {
	return (await import(`./${locale}/${ns}.json`)).default;
}

// A key missing in one locale renders as the raw key in production; keep all locales in lockstep.
describe("messages", () => {
	for (const ns of NAMESPACES) {
		test(`${ns}: every locale has the same keys as en`, async () => {
			const reference = keys(await load("en", ns)).sort();
			for (const locale of routing.locales) {
				expect(keys(await load(locale, ns)).sort()).toEqual(reference);
			}
		});
	}

	test("legal documents parse in every locale with the same sections", async () => {
		for (const doc of ["policy", "terms"] as const) {
			const sizes = await Promise.all(
				routing.locales.map(async (locale) => (await loadLegal(locale, doc)).sections.length),
			);
			expect(new Set(sizes).size).toBe(1);
		}
	});
});
