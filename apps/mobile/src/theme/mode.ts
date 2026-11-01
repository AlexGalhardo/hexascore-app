import { colorScheme } from "nativewind";
import type { PreferenceStore } from "../platform/preferences";

export type ThemePreference = "system" | "light" | "dark";

const KEY = "theme";
const CHOICES: readonly ThemePreference[] = ["system", "light", "dark"];

export function readThemePreference(store: PreferenceStore): ThemePreference {
	const stored = store.get(KEY);
	return CHOICES.find((choice) => choice === stored) ?? "system";
}

export function applyThemePreference(store: PreferenceStore, preference: ThemePreference): void {
	store.set(KEY, preference);
	colorScheme.set(preference);
}

export function restoreThemePreference(store: PreferenceStore): ThemePreference {
	const preference = readThemePreference(store);
	colorScheme.set(preference);
	return preference;
}
