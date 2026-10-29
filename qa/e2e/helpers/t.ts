import { readFileSync } from "node:fs";
import { fileURLToPath } from "node:url";

type Catalogue = { [key: string]: string | Catalogue };

const cache = new Map<string, Catalogue>();

function load(locale: string, file?: string): Catalogue {
	const id = `${locale}/${file ?? ""}`;
	let catalogue = cache.get(id);
	if (!catalogue) {
		const path = file ? `app/accounts/lang/${locale}/${file}.json` : `resources/lang/${locale}.json`;
		catalogue = JSON.parse(readFileSync(fileURLToPath(new URL(`../../../${path}`, import.meta.url)), "utf8")) as Catalogue;
		cache.set(id, catalogue);
	}
	return catalogue;
}

// Translator over the same catalogues the app ships, so specs assert the visible copy in every locale
// without duplicating strings. `server.<file>.<key>` reads the Go-side catalogue app/accounts/lang/<locale>/<file>.json
// (flash messages). `{name}` placeholders are replaced from params; a missing key throws.
export function translator(locale: string): (key: string, params?: Record<string, string | number>) => string {
	return (key, params = {}) => {
		const parts = key.split(".");
		let node: string | Catalogue | undefined = parts[0] === "server" ? load(locale, parts[1]) : load(locale);
		for (const part of parts.slice(parts[0] === "server" ? 2 : 0)) {
			node = typeof node === "object" ? node[part] : undefined;
		}
		if (typeof node !== "string") throw new Error(`missing i18n key ${key} in ${locale}`);
		return node.replace(/\{(\w+)\}/g, (_, name: string) => String(params[name] ?? `{${name}}`));
	};
}

export function escapeRegExp(value: string): string {
	return value.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
}

const LOCALES = ["en", "pt-BR", "es"] as const;

// Matches the copy of `key` in any locale. Form posts redirect to unprefixed URLs (/signup/check-email,
// /two-factor-challenge, /login), which render in the default locale, so a pt-BR or es run cannot assume
// the page language after a post; it still must show the right page.
export function anyLocale(key: string, params?: Record<string, string | number>): RegExp {
	const copies = LOCALES.map((locale) => escapeRegExp(translator(locale)(key, params)));
	return new RegExp(`^(${copies.join("|")})$`);
}
