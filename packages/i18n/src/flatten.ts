// Ports the legacy next-intl catalogues (legacy/apps/web/messages/<locale>/*.json)
// to vukem catalogues (resources/lang/<locale>.json): nested keys, CLDR plural
// maps instead of ICU plural, one key per option instead of ICU select.
// Run: bun packages/i18n/src/flatten.ts <legacy messages dir> <out dir>
import { readdirSync, readFileSync, writeFileSync } from "node:fs";
import { join } from "node:path";

export const LOCALES = ["en", "pt-BR", "es"] as const;

// D-06: per-locale OpenAPI docs, passkeys, PIX/AbacatePay, credit packs, feed
// and follows are out of the rewrite (inventory: task E-013).
const DROPPED_FILES = new Set(["openapi.json"]);
const DROPPED_KEYS = [
	/^account\.passkeys(\.|$)/,
	/^auth\.passkey$/,
	/^auth\.errors\.passkeyFailed$/,
	/^api\.profile\.plan\.credits$/,
	/^billing\..*(pix|abacate|credit|pack)/i,
	/^nav\.feed$/,
	/^billing\.checkout\.(sandboxPay|errors\.(plan_conflict|plan_required|not_found|not_pending))$/,
	/^billing\.reminder(\.|$)/,
	/^docs\.limits\.(creditPacks|colCredits)$/,
	/^docs\.changelog\..*(openapi|locale)/i,
	/^social\.feed(\.|$)/,
	/^social\.profile\.(follow|unfollow|followers|following)$/,
];

type Tree = { [k: string]: string | Tree };

// Plural keys whose ICU variable was renamed to count; callers must pass count.
export const renamed = new Set<string>();

const PLURAL_CATEGORY: Record<string, string> = {
	"=0": "zero",
	zero: "zero",
	one: "one",
	two: "two",
	few: "few",
	many: "many",
	other: "other",
};

// Finds the first "{name, kind, ...}" ICU block and returns its parts.
function findBlock(s: string, kind: "plural" | "select") {
	const m = new RegExp(String.raw`\{\s*(\w+)\s*,\s*${kind}\s*,`).exec(s);
	if (!m) return null;
	let depth = 0;
	for (let i = m.index; i < s.length; i++) {
		if (s[i] === "{") depth++;
		else if (s[i] === "}" && --depth === 0) {
			return {
				name: m[1],
				prefix: s.slice(0, m.index),
				body: s.slice(m.index + m[0].length, i),
				suffix: s.slice(i + 1),
			};
		}
	}
	throw new Error(`unbalanced ICU braces: ${s}`);
}

// Splits "one {a} other {b}" into [["one","a"],["other","b"]].
function options(body: string): [string, string][] {
	const out: [string, string][] = [];
	const re = /\s*([=\w]+)\s*\{/g;
	for (let m = re.exec(body); m; m = re.exec(body)) {
		let depth = 1;
		let i = re.lastIndex;
		for (; i < body.length && depth > 0; i++) {
			if (body[i] === "{") depth++;
			else if (body[i] === "}") depth--;
		}
		out.push([m[1], body.slice(re.lastIndex, i - 1)]);
		re.lastIndex = i;
	}
	return out;
}

export function convert(value: string, key: string): string | Tree {
	const sel = findBlock(value, "select");
	if (sel) {
		const tree: Tree = {};
		for (const [opt, text] of options(sel.body))
			tree[opt] = convert(sel.prefix + text + sel.suffix, key) as string;
		return tree;
	}
	const pl = findBlock(value, "plural");
	if (pl) {
		// vukem picks the plural form from the "count" param, so another ICU
		// variable is renamed to count everywhere in the message.
		const rename = (t: string) =>
			pl.name === "count" ? t : t.replaceAll(`{${pl.name}}`, "{count}");
		if (pl.name !== "count") renamed.add(`${key} (${pl.name} -> count)`);
		const tree: Record<string, string> = {};
		for (const [opt, text] of options(pl.body)) {
			const cat = PLURAL_CATEGORY[opt];
			if (!cat) throw new Error(`${key}: unsupported plural selector ${opt}`);
			tree[cat] = rename(pl.prefix + text.replaceAll("#", "{count}") + pl.suffix);
		}
		if (!tree.other) throw new Error(`${key}: plural without "other"`);
		return tree;
	}
	return value;
}

function walk(node: Tree, path: string, out: Tree, dropped: string[]) {
	for (const [k, v] of Object.entries(node)) {
		const key = path ? `${path}.${k}` : k;
		if (DROPPED_KEYS.some((re) => re.test(key))) {
			dropped.push(key);
			continue;
		}
		if (typeof v === "string") out[k] = convert(v, key);
		else {
			const child: Tree = {};
			walk(v, key, child, dropped);
			if (Object.keys(child).length) out[k] = child;
		}
	}
}

export function buildLocale(dir: string): { tree: Tree; dropped: string[] } {
	const tree: Tree = {};
	const dropped: string[] = [];
	for (const file of readdirSync(dir)
		.filter((f) => f.endsWith(".json") && !DROPPED_FILES.has(f))
		.sort()) {
		walk(JSON.parse(readFileSync(join(dir, file), "utf8")), "", tree, dropped);
	}
	return { tree, dropped };
}

if (import.meta.main) {
	const [src, out] = process.argv.slice(2);
	if (!src || !out) throw new Error("usage: flatten.ts <legacy messages dir> <out dir>");
	for (const locale of LOCALES) {
		const { tree, dropped } = buildLocale(join(src, locale));
		writeFileSync(join(out, `${locale}.json`), `${JSON.stringify(tree, null, 2)}\n`);
		console.log(`${locale}: dropped ${dropped.length} keys`);
	}
	for (const r of renamed) console.log(`renamed plural variable: ${r}`);
}
