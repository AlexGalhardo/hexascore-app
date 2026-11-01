import { existsSync } from "node:fs";
import { join } from "node:path";
import { render, screen } from "@testing-library/react-native";
import config from "../../app.config";
import { fontFamily } from "../theme";
import { RatingInput } from "./RatingInput";
import { TAB_ICONS, TabIcon } from "./TabIcon";

const tailwind = require("../../tailwind.config.js");

describe("fonts", () => {
	type FontObject = { fontFamily: string; fontDefinitions: { path: string; weight: number }[] };
	const plugin = config.plugins?.find(
		(entry) => Array.isArray(entry) && entry[0] === "expo-font",
	) as [string, { android: { fonts: FontObject[] } }] | undefined;
	const families = plugin?.[1].android.fonts ?? [];

	test("expo-font embeds Inter 400/600/700 and Playfair Display 700 from local OFL files", () => {
		const weights = Object.fromEntries(
			families.map((family) => [family.fontFamily, family.fontDefinitions.map((d) => d.weight)]),
		);
		expect(weights).toEqual({ Inter: [400, 600, 700], "Playfair Display": [700] });
		for (const family of families) {
			for (const definition of family.fontDefinitions) {
				expect(existsSync(join(__dirname, "../..", definition.path))).toBe(true);
			}
		}
		expect(existsSync(join(__dirname, "../../assets/fonts/Inter-OFL.txt"))).toBe(true);
		expect(existsSync(join(__dirname, "../../assets/fonts/PlayfairDisplay-OFL.txt"))).toBe(true);
	});

	test("NativeWind font-display and font-text use the embedded families", () => {
		expect(tailwind.theme.extend.fontFamily).toEqual({
			display: [fontFamily.display],
			text: [fontFamily.text],
		});
		expect(fontFamily).toEqual({ display: "Playfair Display", text: "Inter" });
	});
});

describe("icons", () => {
	test("rating stars are icon glyphs: full, half and outline", async () => {
		await render(<RatingInput label="Story" value={3.5} onChange={jest.fn()} locale="en" />);
		const opts = { includeHiddenElements: true };
		expect(screen.getAllByTestId("icon-star-outline", opts)).toHaveLength(10);
		expect(screen.getAllByTestId("icon-star", opts)).toHaveLength(4);
	});

	test("every tab has an outline icon that fills when focused", async () => {
		expect(Object.keys(TAB_ICONS)).toEqual(["index", "[type]", "search", "lists", "profile"]);
		const { rerender } = await render(
			<TabIcon route="search" focused={false} color="#111111" size={24} />,
		);
		expect(screen.getByTestId("icon-search-outline", { includeHiddenElements: true })).toBeTruthy();
		await rerender(<TabIcon route="search" focused color="#111111" size={24} />);
		expect(screen.getByTestId("icon-search", { includeHiddenElements: true }).props.style).toEqual(
			expect.arrayContaining([expect.objectContaining({ color: "#111111", fontSize: 24 })]),
		);
	});
});
