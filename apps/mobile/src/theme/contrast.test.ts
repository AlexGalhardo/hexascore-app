import { readFileSync } from "node:fs";
import { join } from "node:path";
import { contrast } from "../../../../packages/tokens/color";
import { palette } from "./index";

interface TokenFile {
	color: Record<string, Record<"light" | "dark", string>>;
	textPairs: [string, string][];
}

const tokens = JSON.parse(
	readFileSync(join(__dirname, "../../../../packages/tokens/tokens.json"), "utf8"),
) as TokenFile;

const nativewind = require("./read-tokens.cjs") as {
	palette: (mode: "light" | "dark") => Record<string, string>;
};

function hexContrast(a: string, b: string): number {
	const lum = (hex: string): number => {
		const [r, g, bl] = [1, 3, 5].map((i) => {
			const c = Number.parseInt(hex.slice(i, i + 2), 16) / 255;
			return c <= 0.03928 ? c / 12.92 : ((c + 0.055) / 1.055) ** 2.4;
		});
		return 0.2126 * r + 0.7152 * g + 0.0722 * bl;
	};
	const [hi, lo] = [lum(a), lum(b)].sort((x, y) => y - x);
	return (hi + 0.05) / (lo + 0.05);
}

describe("token contrast", () => {
	for (const mode of ["light", "dark"] as const) {
		for (const [fg, bg] of tokens.textPairs) {
			test(`${mode}: ${fg} on ${bg} is at least 4.5:1`, () => {
				expect(contrast(tokens.color[fg][mode], tokens.color[bg][mode])).toBeGreaterThanOrEqual(
					4.5,
				);
			});
		}
	}

	test("neutral skin: text pairs are at least 4.5:1", () => {
		const neutral = require("./neutral.json") as Record<"light" | "dark", Record<string, string>>;
		for (const mode of ["light", "dark"] as const) {
			const colors = palette(mode);
			for (const [fg, bg] of [
				["fg", "bg"],
				["muted", "bg"],
				["muted", "surface"],
				["on-primary", "primary"],
			]) {
				expect(hexContrast(colors[fg], colors[bg])).toBeGreaterThanOrEqual(4.5);
			}
			expect(Object.keys(neutral[mode]).length).toBeGreaterThan(0);
		}
	});

	test("NativeWind theme hex values are the token hex values", () => {
		expect(nativewind.palette("light")).toEqual(palette("light"));
		expect(nativewind.palette("dark")).toEqual(palette("dark"));
	});
});
