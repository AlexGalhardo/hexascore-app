export const THEME_COOKIE = "theme";
export const CONSENT_COOKIE = "consent";
export const ONE_YEAR = 60 * 60 * 24 * 365;

export type Theme = "light" | "dark";
export type Consent = "all" | "essential";

export function setPreferenceCookie(name: string, value: string): void {
	// biome-ignore lint/suspicious/noDocumentCookie: plain preference cookie, Cookie Store API is not available in every browser we support
	document.cookie = `${name}=${value}; path=/; max-age=${ONE_YEAR}; samesite=lax`;
}

export function readPreferenceCookie(name: string): string | undefined {
	return document.cookie
		.split("; ")
		.find((c) => c.startsWith(`${name}=`))
		?.split("=")[1];
}

// Runs before paint to avoid a light/dark flash; must stay dependency-free.
export const themeBootScript = `(function(){try{var m=document.cookie.match(/(?:^|; )theme=(light|dark)/);var t=m?m[1]:(matchMedia('(prefers-color-scheme: dark)').matches?'dark':'light');document.documentElement.dataset.theme=t}catch(e){}})()`;
