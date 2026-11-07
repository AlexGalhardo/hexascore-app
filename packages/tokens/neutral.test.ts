import { readFileSync } from "node:fs";
import { join } from "node:path";
import { describe, expect, it } from "vitest";
import neutral from "../../apps/mobile/src/theme/neutral.json";
import { contrast, oklchToHex } from "./color";

const css = readFileSync(join(import.meta.dirname, "../../resources/css/app.css"), "utf8");

// The neutral zinc theme (ADR 0029, the Android kit in apps/mobile/src/theme/neutral.json) lives twice: as oklch in the web `--hs-*` overrides and
// as hex in the Android kit. This keeps the two from drifting apart.
function block(selector: string, convert: (oklch: string) => string): Record<string, string> {
	const body = new RegExp(`(?:^|\n)${selector.replace(".", ".")} {([^}]*)}`).exec(css)?.[1] ?? "";
	return Object.fromEntries(
		[...body.matchAll(/--hs-([a-z-]+): oklch\(([^)]*)\);/g)].map((m) => [m[1], convert(m[2])]),
	);
}

describe("neutral theme", () => {
	for (const [mode, selector] of [
		["light", ":root"],
		["dark", ".dark"],
	] as const) {
		const web = block(selector, oklchToHex);
		const raw = block(selector, (v) => v);

		it(`${mode}: web and mobile palettes match`, () => {
			for (const [name, hex] of Object.entries(neutral[mode])) {
				expect(web[name], name).toBe(hex);
			}
		});

		it(`${mode}: text meets 4.5:1 and fields 3:1`, () => {
			expect(contrast(raw.fg, raw.bg)).toBeGreaterThanOrEqual(4.5);
			expect(contrast(raw.muted, raw.bg)).toBeGreaterThanOrEqual(4.5);
			expect(contrast(raw.muted, raw.surface)).toBeGreaterThanOrEqual(4.5);
			expect(contrast(raw["on-primary"], raw.primary)).toBeGreaterThanOrEqual(4.5);
			expect(contrast(raw.field, raw["field-bg"])).toBeGreaterThanOrEqual(3);
		});
	}
});

describe("neutral radius scale", () => {
	it("web and mobile use the same sm, md and lg radii", () => {
		for (const name of ["sm", "md", "lg"] as const) {
			expect(css).toContain(`--radius-${name}: ${neutral.radius[name]}px;`);
		}
	});
});
