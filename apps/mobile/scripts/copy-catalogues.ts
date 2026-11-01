import { mkdirSync, readFileSync, writeFileSync } from "node:fs";
import { dirname, join, resolve } from "node:path";
import { fileURLToPath } from "node:url";
import type { Message } from "../src/i18n/format";

export const BUNDLED_NAMESPACES = [
	"common",
	"nav",
	"auth",
	"account",
	"catalog",
	"categories",
	"home",
	"item",
	"itemTypes",
	"lists",
	"ratings",
	"score",
	"search",
	"errors",
] as const;

const PLURAL = new Set(["zero", "one", "two", "few", "many", "other"]);

type Tree = { [key: string]: string | Tree };

export function isPlural(value: Tree): boolean {
	const keys = Object.keys(value);
	return keys.length > 0 && keys.every((key) => PLURAL.has(key));
}

export function flattenNamespace(tree: Tree, prefix: string, out: Record<string, Message>): void {
	for (const [key, value] of Object.entries(tree)) {
		const next = prefix ? `${prefix}.${key}` : key;
		if (typeof value === "string") out[next] = value;
		else if (isPlural(value)) out[next] = value as Record<string, string>;
		else flattenNamespace(value, next, out);
	}
}

export function selectNamespaces(tree: Tree): Record<string, Message> {
	const out: Record<string, Message> = {};
	for (const namespace of BUNDLED_NAMESPACES) {
		const node = tree[namespace];
		if (node && typeof node !== "string") flattenNamespace(node, namespace, out);
	}
	return out;
}

export function loadLocale(repoRoot: string, locale: string): Record<string, Message> {
	const tree = JSON.parse(
		readFileSync(join(repoRoot, "resources/lang", `${locale}.json`), "utf8"),
	) as Tree;
	return selectNamespaces(tree);
}

const LOCALES = ["en", "pt-BR", "es"] as const;

const scriptDir = dirname(fileURLToPath(import.meta.url));

if (process.argv[1] && resolve(process.argv[1]) === resolve(fileURLToPath(import.meta.url))) {
	const repoRoot = join(scriptDir, "../../..");
	const outDir = join(scriptDir, "../src/i18n/catalogues");
	mkdirSync(outDir, { recursive: true });
	for (const locale of LOCALES) {
		const messages = loadLocale(repoRoot, locale);
		writeFileSync(join(outDir, `${locale}.json`), `${JSON.stringify(messages, null, "\t")}\n`);
		console.log(`${locale}: ${Object.keys(messages).length} keys`);
	}
}
