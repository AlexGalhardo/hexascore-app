import { describe, expect, test } from "bun:test";
import { routing } from "@/i18n/routing";
import { buildOpenApiDocument } from "./openapi";

const TEXT_FIELDS = new Set(["description", "summary", "title"]);

/** The document with every human-readable text field removed: paths, schemas and examples only. */
function structure(node: unknown): unknown {
	if (Array.isArray(node)) return node.map(structure);
	if (typeof node !== "object" || node === null) return node;
	return Object.fromEntries(
		Object.entries(node)
			.filter(([key, value]) => !(TEXT_FIELDS.has(key) && typeof value === "string"))
			.map(([key, value]) => [key, structure(value)]),
	);
}

/** `path → text` for every description/summary string. */
function texts(node: unknown, path = ""): Map<string, string> {
	const out = new Map<string, string>();
	if (typeof node !== "object" || node === null) return out;
	for (const [key, value] of Object.entries(node)) {
		const here = `${path}/${key}`;
		if ((key === "description" || key === "summary") && typeof value === "string") out.set(here, value);
		else for (const [p, v] of texts(value, here)) out.set(p, v);
	}
	return out;
}

describe("localized OpenAPI documents", () => {
	const en = buildOpenApiDocument("en");

	for (const locale of routing.locales.filter((l) => l !== "en")) {
		const doc = buildOpenApiDocument(locale);

		test(`${locale}: same paths, schemas and examples as en`, () => {
			expect(structure(doc)).toEqual(structure(en));
		});

		test(`${locale}: every description and summary is translated`, () => {
			const english = texts(en);
			// "OK" is the 200 status text, the same in every language.
			const untranslated = [...texts(doc)].filter(([path, text]) => text === english.get(path) && text !== "OK");
			expect(untranslated).toEqual([]);
		});
	}
});
