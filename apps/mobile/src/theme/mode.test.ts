import { colorScheme } from "nativewind";
import type { PreferenceStore } from "../platform/preferences";
import { applyThemePreference, readThemePreference, restoreThemePreference } from "./mode";

jest.mock("nativewind", () => ({ colorScheme: { set: jest.fn() } }));

function store(initial: Record<string, string> = {}): PreferenceStore {
	const values = new Map(Object.entries(initial));
	return {
		get: (key) => values.get(key) ?? null,
		set: (key, value) => {
			values.set(key, value);
		},
	};
}

describe("theme preference", () => {
	beforeEach(() => jest.mocked(colorScheme.set).mockClear());

	test("defaults to system when nothing or garbage is stored", () => {
		expect(readThemePreference(store())).toBe("system");
		expect(readThemePreference(store({ theme: "sepia" }))).toBe("system");
	});

	test("apply stores the choice and switches NativeWind", () => {
		const prefs = store();
		applyThemePreference(prefs, "dark");
		expect(readThemePreference(prefs)).toBe("dark");
		expect(colorScheme.set).toHaveBeenCalledWith("dark");
	});

	test("restore re-applies the stored choice at start", () => {
		restoreThemePreference(store({ theme: "light" }));
		expect(colorScheme.set).toHaveBeenCalledWith("light");
	});
});
