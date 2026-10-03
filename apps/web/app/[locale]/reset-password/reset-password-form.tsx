"use client";

import { useTranslations } from "next-intl";
import { type FormEvent, useState } from "react";
import { PasswordChecklist } from "@/components/auth/password-checklist";
import { PasswordInput } from "@/components/auth/password-input";
import { buttonClass, Field, Notice } from "@/components/ui/form";
import { Link } from "@/i18n/navigation";
import { authClient } from "@/lib/auth-client";
import { type AuthErrorKey, authErrorKey } from "@/lib/auth-errors";
import { isPasswordValid } from "@/lib/password-rules";

export function ResetPasswordForm({ token }: { token: string }) {
	const t = useTranslations("auth");
	const [password, setPassword] = useState("");
	const [error, setError] = useState<AuthErrorKey | null>(null);
	const [state, setState] = useState<"idle" | "pending" | "done">("idle");

	async function onSubmit(event: FormEvent<HTMLFormElement>): Promise<void> {
		event.preventDefault();
		if (!isPasswordValid(password)) {
			setError("weakPassword");
			return;
		}
		setError(null);
		setState("pending");
		try {
			const { error } = await authClient.resetPassword({ newPassword: password, token });
			if (error) {
				setError(authErrorKey(error));
				setState("idle");
			} else setState("done");
		} catch {
			setError("network");
			setState("idle");
		}
	}

	if (state === "done") {
		return (
			<>
				<Notice tone="success">{t("reset.done")}</Notice>
				<Link href="/login" className={buttonClass("primary")}>
					{t("login.submit")}
				</Link>
			</>
		);
	}

	return (
		<form method="post" onSubmit={onSubmit} className="flex flex-col gap-4" noValidate>
			{error && <Notice tone="error">{t(`errors.${error}`)}</Notice>}
			<Field id="password" label={t("fields.newPassword")}>
				<PasswordInput
					id="password"
					name="password"
					required
					autoComplete="new-password"
					value={password}
					onChange={(e) => setPassword(e.target.value)}
					aria-describedby="password-rules"
				/>
			</Field>
			<PasswordChecklist id="password-rules" value={password} />
			<button type="submit" disabled={state === "pending"} className={buttonClass("primary")}>
				{t("reset.submit")}
			</button>
		</form>
	);
}
