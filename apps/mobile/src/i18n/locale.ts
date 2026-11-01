import { memoryPreferences, type PreferenceStore } from "../platform/preferences";

export const APP_LOCALES = ["en", "pt-BR", "es"] as const;

export type AppLocale = (typeof APP_LOCALES)[number];

const LOCALE_KEY = "hexascore.locale";

export function isAppLocale(value: string): value is AppLocale {
	return (APP_LOCALES as readonly string[]).includes(value);
}

function languageMatch(tag: string): AppLocale | null {
	const lower = tag.trim().replaceAll("_", "-").toLowerCase();
	if (lower === "pt" || lower.startsWith("pt-")) return "pt-BR";
	if (lower === "es" || lower.startsWith("es-")) return "es";
	if (lower === "en" || lower.startsWith("en-")) return "en";
	return null;
}

// First device tag whose language we ship, else en (SDD 10 §8, FR-10-6).
export function matchLocale(tags: readonly string[]): AppLocale {
	for (const tag of tags) {
		const match = languageMatch(tag);
		if (match) return match;
	}
	return "en";
}

export function readLocaleOverride(store: PreferenceStore = memoryPreferences): AppLocale | null {
	const value = store.get(LOCALE_KEY);
	return value && isAppLocale(value) ? value : null;
}

export function writeLocaleOverride(
	locale: AppLocale,
	store: PreferenceStore = memoryPreferences,
): void {
	store.set(LOCALE_KEY, locale);
}

export function resolveLocale(
	deviceTags: readonly string[],
	store: PreferenceStore = memoryPreferences,
): AppLocale {
	return readLocaleOverride(store) ?? matchLocale(deviceTags);
}

export type LocaleStore = {
	get(): AppLocale;
	/** Stores the override and re-renders subscribers (FR-10-6: no restart). */
	set(locale: AppLocale): void;
	subscribe(listener: () => void): () => void;
};

export function createLocaleStore(
	initial: AppLocale,
	store: PreferenceStore = memoryPreferences,
): LocaleStore {
	let current = initial;
	const listeners = new Set<() => void>();
	return {
		get: () => current,
		set(locale) {
			writeLocaleOverride(locale, store);
			if (locale === current) return;
			current = locale;
			for (const listener of listeners) listener();
		},
		subscribe(listener) {
			listeners.add(listener);
			return () => listeners.delete(listener);
		},
	};
}
