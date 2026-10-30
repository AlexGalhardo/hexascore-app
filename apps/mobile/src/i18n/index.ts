import en from "./catalogues/en.json";
import es from "./catalogues/es.json";
import ptBR from "./catalogues/pt-BR.json";
import { type Message, translate } from "./format";
import {
	type AppLocale,
	matchLocale,
	readLocaleOverride,
	resolveLocale,
	writeLocaleOverride,
} from "./locale";

export const catalogues: Record<AppLocale, Record<string, Message>> = {
	en: en as Record<string, Message>,
	"pt-BR": ptBR as Record<string, Message>,
	es: es as Record<string, Message>,
};

export function t(
	locale: AppLocale,
	key: string,
	params: Record<string, string | number> = {},
): string {
	return translate(catalogues[locale], locale, key, params);
}

export type { AppLocale };
export { matchLocale, readLocaleOverride, resolveLocale, writeLocaleOverride };
