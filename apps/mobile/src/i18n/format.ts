// Same rules as resources/js/lib/i18n.ts (SDD 09 §7.1): {name} placeholders,
// CLDR plural objects, Intl.PluralRules, explicit zero wins. Not a second ICU package.
export type Message = string | Record<string, string>;

const fill = (text: string, params: Record<string, string | number>): string =>
	text.replace(/\{(\w+)\}/g, (match: string, name: string) =>
		name in params ? String(params[name]) : match,
	);

// Hermes on Android ships no Intl.PluralRules (it is undefined, and `new` on it throws).
// shortcut: covers en, es and pt-BR only (the app's locales); add a language here before shipping it.
const fallbackCategory = (locale: string, count: number): string =>
	locale.toLowerCase().startsWith("pt")
		? count >= 0 && count < 2
			? "one"
			: "other"
		: count === 1
			? "one"
			: "other";

export function translate(
	messages: Record<string, Message>,
	locale: string,
	key: string,
	params: Record<string, string | number> = {},
): string {
	const message = messages[key];
	if (message === undefined) return key;
	if (typeof message === "string") return fill(message, params);

	const count = params.count;
	if (typeof count !== "number") return fill(message.other ?? key, params);
	if (count === 0 && message.zero !== undefined) return fill(message.zero, params);
	const category =
		typeof Intl.PluralRules === "function"
			? new Intl.PluralRules(locale).select(count)
			: fallbackCategory(locale, count);
	return fill(message[category] ?? message.other ?? key, params);
}
