import type { PreferenceStore } from "../platform/preferences";
import { matchLocale, resolveLocale, writeLocaleOverride } from "./locale";

function store(): PreferenceStore {
	const data = new Map<string, string>();
	return {
		get: (key) => data.get(key) ?? null,
		set: (key, value) => {
			data.set(key, value);
		},
	};
}

test("pt-PT matches pt-BR, es-MX matches es, fr falls back to en", () => {
	expect(matchLocale(["pt-PT"])).toBe("pt-BR");
	expect(matchLocale(["es-MX"])).toBe("es");
	expect(matchLocale(["fr"])).toBe("en");
});

test("the first supported language wins", () => {
	expect(matchLocale(["fr", "es-MX"])).toBe("es");
	expect(matchLocale(["en-US"])).toBe("en");
	expect(matchLocale(["pt-BR"])).toBe("pt-BR");
});

test("an in-app override beats the device list", () => {
	const prefs = store();
	writeLocaleOverride("es", prefs);
	expect(resolveLocale(["fr"], prefs)).toBe("es");
	writeLocaleOverride("en", prefs);
	expect(resolveLocale(["pt-PT"], prefs)).toBe("en");
});
