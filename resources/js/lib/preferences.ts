export const CONSENT_COOKIE = "consent";

export const CONSENT_EVENT = "hexascore:consent";

export const COOKIE_SETTINGS_EVENT = "hexascore:cookie-settings";

export type Consent = "all" | "essential";

declare global {
	interface Window {
		// Defined by the {{analytics}} GA4 bootstrap in app.html only when GA_MEASUREMENT_ID is set.
		vukemAnalytics?: { consent: (granted: boolean) => void };
	}
}

// applyAnalyticsConsent passes a stored choice to GA4 consent mode: "all" grants, "essential" denies, no choice
// leaves GA denied (it loads nothing before a grant). Umami is cookieless and does not depend on it.
export function applyAnalyticsConsent(value: string | null): void {
	if (value === "all" || value === "essential") window.vukemAnalytics?.consent(value === "all");
}

// openCookieSettings shows the cookie banner again so the visitor can change the choice.
export function openCookieSettings(): void {
	window.dispatchEvent(new Event(COOKIE_SETTINGS_EVENT));
}

export function readPreferenceCookie(name: string): string | null {
	const match = document.cookie.split("; ").find((part) => part.startsWith(`${name}=`));
	if (!match) return null;
	return decodeURIComponent(match.slice(name.length + 1));
}

export function setPreferenceCookie(name: string, value: string): void {
	// biome-ignore lint/suspicious/noDocumentCookie: a plain preference cookie, same as the theme toggle
	document.cookie = `${name}=${encodeURIComponent(value)}; path=/; max-age=31536000; samesite=lax`;
	if (name === CONSENT_COOKIE) {
		applyAnalyticsConsent(value);
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
