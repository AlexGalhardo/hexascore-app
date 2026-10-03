"use client";

import { useLocale, useTranslations } from "next-intl";
import { type FormEvent, useState } from "react";
import { GoogleButton } from "@/components/auth/google-button";
import { PasswordChecklist } from "@/components/auth/password-checklist";
import { PasswordInput } from "@/components/auth/password-input";
import { buttonClass, Divider, Field, Input, Notice } from "@/components/ui/form";
import { getPathname, Link } from "@/i18n/navigation";
import { authClient } from "@/lib/auth-client";
import { type AuthErrorKey, authErrorKey } from "@/lib/auth-errors";
import { emailSchema, NICKNAME_MAX, nameSchema, nicknameSchema, passwordSchema } from "@/lib/validation";

type FieldName = "nickname" | "name" | "email" | "password";

const ERROR_KEY: Record<FieldName, string> = {
	nickname: "invalidNickname",
	name: "invalidName",
	email: "invalidEmail",
	password: "weakPassword",
};

export function SignupForm({ googleEnabled }: { googleEnabled: boolean }) {
	const t = useTranslations("auth");
	const locale = useLocale();
	const [nickname, setNickname] = useState("");
	const [password, setPassword] = useState("");
	const [invalid, setInvalid] = useState<Set<FieldName>>(new Set());
	const [error, setError] = useState<AuthErrorKey | null>(null);
	const [pending, setPending] = useState(false);
	const [sentTo, setSentTo] = useState<string | null>(null);

	async function onSubmit(event: FormEvent<HTMLFormElement>): Promise<void> {
		event.preventDefault();
		const form = new FormData(event.currentTarget);
		const values = {
			nickname: String(form.get("nickname") ?? "").trim(),
			name: String(form.get("name") ?? "").trim(),
			email: String(form.get("email") ?? "").trim(),
			password: String(form.get("password") ?? ""),
		};
		const checks: Record<FieldName, boolean> = {
			nickname: nicknameSchema.safeParse(values.nickname).success,
			name: nameSchema.safeParse(values.name).success,
			email: emailSchema.safeParse(values.email).success,
			password: passwordSchema.safeParse(values.password).success,
		};
		const bad = new Set((Object.keys(checks) as FieldName[]).filter((k) => !checks[k]));
		setInvalid(bad);
		if (bad.size > 0) return;

		setPending(true);
		setError(null);
		try {
			const { error } = await authClient.signUp.email({
				email: values.email,
				password: values.password,
				name: values.name,
				username: values.nickname,
				callbackURL: getPathname({ href: "/verify-email", locale }),
			});
			if (error) setError(authErrorKey(error));
			else setSentTo(values.email);
		} catch {
			setError("network");
		} finally {
			setPending(false);
		}
	}

	if (sentTo) {
		return (
			<Notice tone="success">
				<p className="font-semibold">{t("signup.checkTitle")}</p>
				<p className="mt-1">{t("signup.checkBody", { email: sentTo })}</p>
			</Notice>
		);
	}

	const fieldError = (name: FieldName): string | undefined =>
		invalid.has(name) ? t(`errors.${ERROR_KEY[name]}`) : undefined;
	const describedBy = (name: FieldName, extra?: string): string =>
		[extra, invalid.has(name) ? `${name}-error` : null].filter(Boolean).join(" ");

	return (
		<div className="flex flex-col gap-4">
			<form onSubmit={onSubmit} className="flex flex-col gap-4" noValidate>
				{error && <Notice tone="error">{t(`errors.${error}`)}</Notice>}
				<div className="flex items-end gap-3">
					<div
						role="img"
						aria-label={t("signup.avatar")}
						className="grid size-14 shrink-0 place-items-center rounded-full bg-surface font-serif text-2xl font-bold ring-1 ring-border"
					>
						{nickname.trim().charAt(0).toUpperCase() || "?"}
					</div>
					<div className="flex-1">
						<Field id="nickname" label={t("fields.nickname")} error={fieldError("nickname")}>
							<Input
								id="nickname"
								name="nickname"
								required
								autoComplete="username"
								maxLength={NICKNAME_MAX}
								value={nickname}
								onChange={(e) => setNickname(e.target.value)}
								aria-invalid={invalid.has("nickname")}
								aria-describedby={describedBy("nickname", "nickname-hint")}
							/>
						</Field>
					</div>
				</div>
				<p id="nickname-hint" className="-mt-2 text-sm text-muted">
					{t("signup.nicknameHint")}
				</p>
				<Field id="name" label={t("fields.name")} error={fieldError("name")}>
					<Input
						id="name"
						name="name"
						required
						autoComplete="name"
						maxLength={100}
						aria-invalid={invalid.has("name")}
						aria-describedby={describedBy("name")}
					/>
				</Field>
				<Field id="email" label={t("fields.email")} error={fieldError("email")}>
					<Input
						id="email"
						name="email"
						type="email"
						required
						autoComplete="email"
						aria-invalid={invalid.has("email")}
						aria-describedby={describedBy("email")}
					/>
				</Field>
				<Field id="password" label={t("fields.password")} error={fieldError("password")}>
					<PasswordInput
						id="password"
						name="password"
						required
						autoComplete="new-password"
						value={password}
						onChange={(e) => setPassword(e.target.value)}
						aria-invalid={invalid.has("password")}
						aria-describedby={describedBy("password", "password-rules")}
					/>
				</Field>
				<PasswordChecklist id="password-rules" value={password} />
				<p className="text-sm text-muted">
					{t.rich("signup.agree", {
						terms: (chunks) => (
							<Link href="/terms" className="underline underline-offset-4">
								{chunks}
							</Link>
						),
						privacy: (chunks) => (
							<Link href="/policy" className="underline underline-offset-4">
								{chunks}
							</Link>
						),
					})}
				</p>
				<button type="submit" disabled={pending} className={buttonClass("primary")}>
					{t("signup.submit")}
				</button>
			</form>
			{googleEnabled && (
				<>
					<Divider label={t("or")} />
					<GoogleButton callbackURL={getPathname({ href: "/profile", locale })} />
				</>
			)}
		</div>
	);
}
