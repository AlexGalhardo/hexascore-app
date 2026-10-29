import { readFileSync } from "node:fs";
import { describe, expect, test } from "vitest";

import { LOCALES } from "./flatten";

type Tree = { [k: string]: string | Tree };

function leaves(tree: Tree, prefix = ""): string[] {
	return Object.entries(tree).flatMap(([k, v]) => {
		const key = prefix ? `${prefix}.${k}` : k;
		return typeof v === "string" ? [key] : leaves(v, key);
	});
}

const load = (locale: string): Tree =>
	JSON.parse(readFileSync(`resources/lang/${locale}.json`, "utf8")) as Tree;

describe("catalogue parity", () => {
	const en = new Set(leaves(load("en")));
	for (const locale of LOCALES.filter((l) => l !== "en")) {
		test(`every en key exists in ${locale}`, () => {
			const other = new Set(leaves(load(locale)));
			expect([...en].filter((k) => !other.has(k))).toEqual([]);
		});
	}
});

// JSON.parse keeps the last of two equal keys, so a duplicated block silently drops the first one's keys.
function duplicateKeys(text: string): string[] {
	const path: string[] = [];
	const seen = new Set<string>();
	const dups: string[] = [];
	for (const line of text.split("\n")) {
		const m = /^( *)"([^"]+)":/.exec(line);
		if (!m) continue;
		path.length = m[1].length / 2 - 1;
		const key = [...path, m[2]].join(".");
		if (seen.has(key)) dups.push(key);
		seen.add(key);
		path.push(m[2]);
	}
	return dups;
}

describe("catalogue keys", () => {
	for (const locale of LOCALES) {
		test(`${locale} has no duplicate keys`, () => {
			expect(duplicateKeys(readFileSync(`resources/lang/${locale}.json`, "utf8"))).toEqual(
				[],
			);
		});
	}
});
