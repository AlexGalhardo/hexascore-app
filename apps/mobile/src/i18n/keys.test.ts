import { join } from "node:path";
import { loadLocale } from "../../scripts/copy-catalogues";
import { translate } from "./format";
import { catalogues } from "./index";

const repoRoot = join(__dirname, "../../../..");
const excluded = ["api", "billing", "docs", "openapi", "contact", "legal", "social"];

test("every bundled en key exists in pt-BR and es", () => {
	const en = loadLocale(repoRoot, "en");
	const pt = loadLocale(repoRoot, "pt-BR");
	const es = loadLocale(repoRoot, "es");
	expect(Object.keys(en).length).toBeGreaterThan(0);
	expect(Object.keys(en).filter((key) => !(key in pt))).toEqual([]);
	expect(Object.keys(en).filter((key) => !(key in es))).toEqual([]);
});

test("committed catalogues are the build-time copy", () => {
	expect(catalogues.en).toEqual(loadLocale(repoRoot, "en"));
	expect(catalogues["pt-BR"]).toEqual(loadLocale(repoRoot, "pt-BR"));
	expect(catalogues.es).toEqual(loadLocale(repoRoot, "es"));
});

test("excluded namespaces are not bundled", () => {
	for (const namespace of excluded) {
		expect(
			Object.keys(catalogues.en).some(
				(key) => key === namespace || key.startsWith(`${namespace}.`),
			),
		).toBe(false);
	}
});

test("plural messages format with Intl.PluralRules", () => {
	const key = Object.keys(catalogues.en).find((name) => typeof catalogues.en[name] === "object");
	expect(key).toBeDefined();
	const one = translate(catalogues.en, "en", key ?? "", { count: 1 });
	const many = translate(catalogues.en, "en", key ?? "", { count: 2 });
	expect(one).not.toBe(key);
	expect(many).not.toBe(one);
});

test("plural messages still format when the runtime has no Intl.PluralRules (Hermes)", () => {
	const original = Intl.PluralRules;
	Object.defineProperty(Intl, "PluralRules", { value: undefined, configurable: true });
	try {
		const messages = { items: { one: "{count} item", other: "{count} items" } };
		expect(translate(messages, "en", "items", { count: 1 })).toBe("1 item");
		expect(translate(messages, "en", "items", { count: 2 })).toBe("2 items");
		expect(translate(messages, "pt-BR", "items", { count: 0 })).toBe("0 item");
		expect(translate(messages, "pt-BR", "items", { count: 2 })).toBe("2 items");
	} finally {
		Object.defineProperty(Intl, "PluralRules", { value: original, configurable: true });
	}
});
