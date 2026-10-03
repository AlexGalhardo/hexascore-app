"use client";

import { useTranslations } from "next-intl";
import { type FormEvent, useState } from "react";
import { setPassword } from "@/actions/account";
import { PasswordChecklist } from "@/components/auth/password-checklist";
import { PasswordInput } from "@/components/auth/password-input";
import { buttonClass, Field, Notice } from "@/components/ui/form";
import { authClient } from "@/lib/auth-client";
import { authErrorKey } from "@/lib/auth-errors";
import { isPasswordValid } from "@/lib/password-rules";

type Feedback = { tone: "error" | "success"; text: string } | null;

// Accounts with a password change it (Better Auth client); Google/passkey-only accounts set one (Server Action).
export function PasswordSection({ hasPassword }: { hasPassword: boolean }) {
	const t = useTranslations("account");
	const tAuth = useTranslations("auth");
	const tAction = useTranslations("actions");
	const [password, setPasswordValue] = useState("");
	const [feedback, setFeedback] = useState<Feedback>(null);
	const [pending, setPending] = useState(false);

	async function onSubmit(event: FormEvent<HTMLFormElement>): Promise<void> {
		event.preventDefault();
		const form = event.currentTarget;
		const data = new FormData(form);
		if (!isPasswordValid(password)) {
			setFeedback({ tone: "error", text: tAuth("errors.weakPassword") });
			return;
		}
		setPending(true);
		setFeedback(null);
		try {
			if (hasPassword) {
				const { error } = await authClient.changePassword({
					currentPassword: String(data.get("currentPassword") ?? ""),
					newPassword: password,
					revokeOtherSessions: data.get("revokeOthers") === "on",
				});
				if (error) {
					setFeedback({ tone: "error", text: tAuth(`errors.${authErrorKey(error)}`) });
					return;
				}
			} else {
				const result = await setPassword(password);
				if (!result.ok) {
					setFeedback({ tone: "error", text: tAction(result.error) });
					return;
				}
			}
			form.reset();
			setPasswordValue("");
			setFeedback({ tone: "success", text: t("password.done") });
		} catch {
			setFeedback({ tone: "error", text: tAuth("errors.network") });
		} finally {
			setPending(false);
		}
	}

	return (
		<form method="post" onSubmit={onSubmit} className="flex flex-col gap-4" noValidate>
			{!hasPassword && <p className="text-sm text-muted">{t("password.setBody")}</p>}
			{feedback && <Notice tone={feedback.tone}>{feedback.text}</Notice>}
			{hasPassword && (
				<Field id="current-password" label={tAuth("fields.currentPassword")}>
					<PasswordInput
						id="current-password"
						name="currentPassword"
						required
						autoComplete="current-password"
					/>
				</Field>
			)}
			<Field id="new-password" label={tAuth("fields.newPassword")}>
				<PasswordInput
					id="new-password"
					name="newPassword"
					required
					autoComplete="new-password"
					value={password}
					onChange={(e) => setPasswordValue(e.target.value)}
					aria-describedby="new-password-rules"
				/>
			</Field>
			<PasswordChecklist id="new-password-rules" value={password} />
			{hasPassword && (
				<label className="flex min-h-11 items-center gap-2 text-sm">
					<input type="checkbox" name="revokeOthers" defaultChecked className="size-5" />
					{t("password.revokeOthers")}
				</label>
			)}
			<button type="submit" disabled={pending} className={buttonClass("secondary", "self-start")}>
				{hasPassword ? t("password.change") : t("password.set")}
			</button>
		</form>
	);
}
