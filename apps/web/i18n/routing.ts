import { defineRouting } from "next-intl/routing";

export const routing = defineRouting({
	locales: ["en", "pt-BR", "es"],
	defaultLocale: "en",
	localePrefix: { mode: "as-needed", prefixes: { "pt-BR": "/pt-br", es: "/es" } },
	localeCookie: { name: "locale", maxAge: 60 * 60 * 24 * 365 },
});

export type Locale = (typeof routing.locales)[number];
