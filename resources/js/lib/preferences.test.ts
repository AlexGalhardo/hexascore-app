import { expect, test } from "vitest";

import {
	CONSENT_COOKIE,
	onConsentChange,
	readPreferenceCookie,
	setPreferenceCookie,
} from "./preferences";

test("setting consent notifies listeners in the same tab", () => {
	// biome-ignore lint/suspicious/noDocumentCookie: jsdom has no Cookie Store API
	document.cookie = `${CONSENT_COOKIE}=; path=/; max-age=0`;
	const seen: Array<string | null> = [];
	const stop = onConsentChange((value) => {
		seen.push(value);
	});

	setPreferenceCookie(CONSENT_COOKIE, "all");

	expect(readPreferenceCookie(CONSENT_COOKIE)).toBe("all");
	expect(seen).toEqual(["all"]);
	stop();
	setPreferenceCookie(CONSENT_COOKIE, "essential");
	expect(seen).toEqual(["all"]);
});
