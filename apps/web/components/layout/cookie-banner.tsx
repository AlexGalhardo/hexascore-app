"use client";

import { useEffect, useState } from "react";
import { CONSENT_COOKIE, type Consent, readPreferenceCookie, setPreferenceCookie } from "@/lib/preferences";

type Props = { text: string; accept: string; decline: string };

export function CookieBanner({ text, accept, decline }: Props) {
	const [open, setOpen] = useState(false);

	useEffect(() => {
		setOpen(!readPreferenceCookie(CONSENT_COOKIE));
	}, []);

	function choose(value: Consent): void {
		setPreferenceCookie(CONSENT_COOKIE, value);
		setOpen(false);
	}

	if (!open) return null;

	return (
		<section
			aria-label="Cookies"
			className="fixed inset-x-0 bottom-0 z-50 border-t border-border bg-surface p-4 shadow-lg"
		>
			<div className="mx-auto flex max-w-6xl flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
				<p className="text-sm">{text}</p>
				<div className="flex shrink-0 gap-2">
					<button
						type="button"
						onClick={() => choose("essential")}
						className="min-h-11 rounded-md border border-border px-4 text-sm"
					>
						{decline}
					</button>
					<button
						type="button"
						onClick={() => choose("all")}
						className="min-h-11 rounded-md bg-fg px-4 text-sm text-bg"
					>
						{accept}
					</button>
				</div>
			</div>
		</section>
	);
}
