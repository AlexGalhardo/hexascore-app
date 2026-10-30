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

	test("NativeWind theme hex values are the token hex values", () => {
		expect(nativewind.palette("light")).toEqual(palette("light"));
		expect(nativewind.palette("dark")).toEqual(palette("dark"));
	});
});
