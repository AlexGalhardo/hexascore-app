export const CONSENT_COOKIE = "consent";

export const CONSENT_EVENT = "hexascore:consent";

export type Consent = "all" | "essential";

export function readPreferenceCookie(name: string): string | null {
	const match = document.cookie.split("; ").find((part) => part.startsWith(`${name}=`));
	if (!match) return null;
	return decodeURIComponent(match.slice(name.length + 1));
}

export function setPreferenceCookie(name: string, value: string): void {
	// biome-ignore lint/suspicious/noDocumentCookie: a plain preference cookie, same as the theme toggle
	document.cookie = `${name}=${encodeURIComponent(value)}; path=/; max-age=31536000; samesite=lax`;
	if (name === CONSENT_COOKIE) {
		window.dispatchEvent(new CustomEvent(CONSENT_EVENT, { detail: value }));
	}
}

export function onConsentChange(listener: (value: string | null) => void): () => void {
	const handler = (): void => {
		listener(readPreferenceCookie(CONSENT_COOKIE));
	};
	window.addEventListener(CONSENT_EVENT, handler);
	return () => window.removeEventListener(CONSENT_EVENT, handler);
}
