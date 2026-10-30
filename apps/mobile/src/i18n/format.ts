// Same rules as resources/js/lib/i18n.ts (SDD 09 §7.1): {name} placeholders,
// CLDR plural objects, Intl.PluralRules, explicit zero wins. Not a second ICU package.
export type Message = string | Record<string, string>;

const fill = (text: string, params: Record<string, string | number>): string =>
	text.replace(/\{(\w+)\}/g, (match: string, name: string) =>
		name in params ? String(params[name]) : match,
	);

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
	const category = new Intl.PluralRules(locale).select(count);
	return fill(message[category] ?? message.other ?? key, params);
}
