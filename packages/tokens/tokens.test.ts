import { readdirSync, readFileSync } from "node:fs";
import { join } from "node:path";
import { describe, expect, it } from "vitest";
import { renderCss, renderTs } from "./build";
import { contrast, oklchToHex } from "./color";
import tokens from "./tokens.json";

type Name = keyof typeof tokens.color;

describe("tokens", () => {
	for (const mode of ["light", "dark"] as const) {
		for (const [fg, bg] of tokens.textPairs as [Name, Name][]) {
			it(`${mode}: ${fg} on ${bg} meets WCAG AA (4.5:1)`, () => {
				expect(
					contrast(tokens.color[fg][mode], tokens.color[bg][mode]),
				).toBeGreaterThanOrEqual(4.5);
			});
		}
	}

	it("converts OKLCH to sRGB hex", () => {
		expect(oklchToHex("1 0 0")).toBe("#ffffff");
		expect(oklchToHex("0 0 0")).toBe("#000000");
		expect(oklchToHex("0.628 0.2577 29.23")).toBe("#ff0000");
	});

	it("committed outputs are up to date", () => {
		expect(readFileSync(join(import.meta.dirname, "tokens.css"), "utf8")).toBe(renderCss());
		expect(readFileSync(join(import.meta.dirname, "tokens.ts"), "utf8")).toBe(renderTs());
	});

	it("no hardcoded indigo-* classes remain in resources/js", () => {
		const root = join(import.meta.dirname, "..", "..", "resources", "js");
		const hits = (readdirSync(root, { recursive: true, encoding: "utf8" }) as string[])
			.filter((f) => /\.(tsx?|css)$/.test(f))
			.filter((f) => /indigo-/.test(readFileSync(join(root, f), "utf8")));
		expect(hits).toEqual([]);
	});
});
