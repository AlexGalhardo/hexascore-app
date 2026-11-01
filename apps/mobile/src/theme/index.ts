import { colors, fontSize, radius, spacing, touchTarget } from "../../../../packages/tokens/tokens";

export type { ThemePreference } from "./mode";
export { applyThemePreference, readThemePreference, restoreThemePreference } from "./mode";
export { colors, fontSize, radius, spacing, touchTarget };

export type ThemeName = keyof typeof colors;

// Embedded on Android by the expo-font config plugin (app.config.ts); keep tailwind.config.js in step.
export const fontFamily = {
	display: "Playfair Display",
	text: "Inter",
} as const;

export function palette(mode: ThemeName): Record<string, string> {
	return Object.fromEntries(Object.entries(colors[mode]).map(([name, value]) => [name, value.hex]));
}

export function spacingPx(step: (typeof spacing)[number]): number {
	return step * 4;
}
