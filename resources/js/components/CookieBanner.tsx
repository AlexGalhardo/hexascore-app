import { usePage } from "@inertiajs/react";
import { useEffect, useRef, useState } from "react";
import {
	applyAnalyticsConsent,
	CONSENT_COOKIE,
	COOKIE_SETTINGS_EVENT,
	type Consent,
	readPreferenceCookie,
	setPreferenceCookie,
} from "../lib/preferences";
import { useT } from "../lib/useT";
import type { SharedProps } from "../types";

// CookieBanner records the LGPD choice in the first-party "consent" cookie: "all" allows YouTube embeds and GA4,
// "essential" refuses both (ADR 0023). The footer's "Cookie settings" reopens it.
export default function CookieBanner() {
	const t = useT();
	const consent = usePage<SharedProps>().props.consent;
	const [open, setOpen] = useState(consent !== "all" && consent !== "essential");
	const first = useRef<HTMLButtonElement>(null);
	const reopened = useRef(false);

	useEffect(() => {
		// Replays the stored choice on every full load: GA4 starts denied each time.
		applyAnalyticsConsent(readPreferenceCookie(CONSENT_COOKIE) ?? consent ?? null);
		const show = (): void => {
			reopened.current = true;
			setOpen(true);
		};
		window.addEventListener(COOKIE_SETTINGS_EVENT, show);
		return () => window.removeEventListener(COOKIE_SETTINGS_EVENT, show);
	}, [consent]);

	useEffect(() => {
		if (open && reopened.current) first.current?.focus();
	}, [open]);

	function choose(value: Consent): void {
		setPreferenceCookie(CONSENT_COOKIE, value);
		setOpen(false);
	}

	if (!open) return null;

	return (
		<>
			{/* Keeps the last controls of the page reachable while the fixed banner is shown. */}
			<div aria-hidden="true" className="h-36 sm:h-20" />
			<section
				aria-label={t("cookies.label")}
				className="fixed inset-x-0 bottom-0 z-50 border-t border-border bg-surface p-4 shadow-lg"
			>
				<div className="mx-auto flex max-w-6xl flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
					<p className="text-sm">{t("cookies.text")}</p>
					<div className="flex shrink-0 gap-2">
						<button
							ref={first}
							type="button"
							onClick={() => choose("essential")}
							className="min-h-11 rounded-md border border-border px-4 text-sm focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fg"
						>
							{t("cookies.decline")}
						</button>
						<button
							type="button"
							onClick={() => choose("all")}
							className="min-h-11 rounded-md bg-fg px-4 text-sm text-bg focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fg"
						>
							{t("cookies.accept")}
						</button>
					</div>
				</div>
			</section>
		</>
	);
}
