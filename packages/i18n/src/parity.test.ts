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
