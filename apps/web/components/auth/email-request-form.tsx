"use client";

import { useLocale, useTranslations } from "next-intl";
import { type FormEvent, useState } from "react";
import { buttonClass, Field, Input, Notice } from "@/components/ui/form";
import { getPathname } from "@/i18n/navigation";
import { authClient } from "@/lib/auth-client";
import { type AuthErrorKey, authErrorKey } from "@/lib/auth-errors";
import { emailSchema } from "@/lib/validation";

type Props = { kind: "verify" | "recover"; defaultEmail?: string | undefined };

// Shared by "resend verification" and "recover password". Both always show the same confirmation
// whether or not the email exists, so the form cannot be used to enumerate accounts.
export function EmailRequestForm({ kind, defaultEmail }: Props) {
	const t = useTranslations("auth");
	const locale = useLocale();
	const [state, setState] = useState<"idle" | "pending" | "sent">("idle");
	const [error, setError] = useState<AuthErrorKey | "invalidEmail" | null>(null);

	async function onSubmit(event: FormEvent<HTMLFormElement>): Promise<void> {
		event.preventDefault();
		const email = String(new FormData(event.currentTarget).get("email") ?? "").trim();
		if (!emailSchema.safeParse(email).success) {
			setError("invalidEmail");
			return;
		}
		setError(null);
		setState("pending");
		try {
			const { error } =
				kind === "verify"
					? await authClient.sendVerificationEmail({
							email,
							callbackURL: getPathname({ href: "/verify-email", locale }),
						})
					: await authClient.requestPasswordReset({
							email,
							redirectTo: getPathname({ href: "/reset-password", locale }),
						});
			const key = error ? authErrorKey(error) : null;
			if (key === "rateLimited" || key === "network" || key === "bot") {
				setError(key);
				setState("idle");
			} else {
				setState("sent");
			}
		} catch {
			setError("network");
			setState("idle");
		}
	}

	if (state === "sent") {
		return <Notice tone="success">{kind === "verify" ? t("verify.resent") : t("recover.sent")}</Notice>;
	}

	return (
		<form onSubmit={onSubmit} className="flex flex-col gap-4" noValidate>
			{error && <Notice tone="error">{t(`errors.${error}`)}</Notice>}
			<Field id="email" label={t("fields.email")}>
				<Input
					id="email"
					name="email"
					type="email"
					required
					autoComplete="email"
					defaultValue={defaultEmail}
					aria-invalid={error === "invalidEmail"}
				/>
			</Field>
			<button type="submit" disabled={state === "pending"} className={buttonClass("primary")}>
				{kind === "verify" ? t("verify.resend") : t("recover.submit")}
			</button>
		</form>
	);
}
