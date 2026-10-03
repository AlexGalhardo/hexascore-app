"use client";

import { useTranslations } from "next-intl";
import { useState } from "react";
import { buttonClass } from "@/components/ui/form";
import { authClient } from "@/lib/auth-client";

export function GoogleButton({ callbackURL }: { callbackURL: string }) {
	const t = useTranslations("auth");
	const [pending, setPending] = useState(false);

	async function signIn(): Promise<void> {
		setPending(true);
		// Success navigates away to Google; only a failure comes back here.
		const { error } = await authClient.signIn.social({ provider: "google", callbackURL });
		if (error) setPending(false);
	}

	return (
		<button type="button" onClick={signIn} disabled={pending} className={buttonClass("secondary", "w-full")}>
			<svg aria-hidden="true" viewBox="0 0 24 24" className="size-5">
				<path
					fill="#4285F4"
					d="M23.5 12.3c0-.8-.1-1.6-.2-2.3H12v4.4h6.5a5.6 5.6 0 0 1-2.4 3.6v3h3.9c2.3-2.1 3.5-5.2 3.5-8.7z"
				/>
				<path
					fill="#34A853"
					d="M12 24c3.2 0 6-1.1 8-2.9l-3.9-3c-1.1.7-2.5 1.2-4.1 1.2-3.1 0-5.8-2.1-6.7-5H1.3v3.1A12 12 0 0 0 12 24z"
				/>
				<path fill="#FBBC05" d="M5.3 14.3a7.2 7.2 0 0 1 0-4.6V6.6H1.3a12 12 0 0 0 0 10.8l4-3.1z" />
				<path
					fill="#EA4335"
					d="M12 4.8c1.8 0 3.3.6 4.6 1.8l3.4-3.4A12 12 0 0 0 1.3 6.6l4 3.1c.9-2.9 3.6-4.9 6.7-4.9z"
				/>
			</svg>
			{t("google")}
		</button>
	);
}
