"use client";

import { useLocale, useTranslations } from "next-intl";
import { type FormEvent, useState } from "react";
import { GoogleButton } from "@/components/auth/google-button";
import { PasswordInput } from "@/components/auth/password-input";
import { buttonClass, Divider, Field, Input, Notice } from "@/components/ui/form";
import { getPathname, Link } from "@/i18n/navigation";
import { authClient } from "@/lib/auth-client";
import { type AuthClientError, type AuthErrorKey, authErrorKey } from "@/lib/auth-errors";

type Props = { next: string; googleEnabled: boolean };
type Step = "credentials" | "totp" | "backup";

export function LoginForm({ next, googleEnabled }: Props) {
	const t = useTranslations("auth");
	const locale = useLocale();
	const [step, setStep] = useState<Step>("credentials");
	const [error, setError] = useState<AuthErrorKey | null>(null);
	const [pending, setPending] = useState(false);
	const [unverifiedEmail, setUnverifiedEmail] = useState<string | null>(null);
	const [resent, setResent] = useState(false);

	// Full navigation so the server-rendered navbar and banners pick up the new session.
	const finish = (): void => window.location.assign(next);

	async function run(call: () => Promise<{ error: AuthClientError }>): Promise<boolean> {
		setPending(true);
		setError(null);
		try {
			const { error } = await call();
			if (error) {
				setError(authErrorKey(error));
				return false;
			}
			return true;
		} catch {
			setError("network");
			return false;
		} finally {
			setPending(false);
		}
	}

	async function onCredentials(event: FormEvent<HTMLFormElement>): Promise<void> {
		event.preventDefault();
		const form = new FormData(event.currentTarget);
		const identifier = String(form.get("identifier") ?? "").trim();
		const password = String(form.get("password") ?? "");
		const isEmail = identifier.includes("@");
		let twoFactor = false;
		setUnverifiedEmail(null);
		setResent(false);
		const success = await run(async () => {
			const result = isEmail
				? await authClient.signIn.email({ email: identifier, password })
				: await authClient.signIn.username({ username: identifier, password });
			if (result.error?.code === "EMAIL_NOT_VERIFIED" && isEmail) setUnverifiedEmail(identifier);
			twoFactor = Boolean(result.data && "twoFactorRedirect" in result.data && result.data.twoFactorRedirect);
			return result;
		});
		if (!success) return;
		if (twoFactor) setStep("totp");
		else finish();
	}

	async function onCode(event: FormEvent<HTMLFormElement>): Promise<void> {
		event.preventDefault();
		const form = new FormData(event.currentTarget);
		const code = String(form.get("code") ?? "").trim();
		const trustDevice = form.get("trustDevice") === "on";
		const success = await run(() =>
			step === "totp"
				? authClient.twoFactor.verifyTotp({ code, trustDevice })
				: authClient.twoFactor.verifyBackupCode({ code, trustDevice }),
		);
		if (success) finish();
	}

	async function onPasskey(): Promise<void> {
		const success = await run(() => authClient.signIn.passkey());
		if (success) finish();
		// A cancelled browser prompt comes back as a generic error; say what actually happened.
		else setError((e) => (e === "generic" ? "passkeyFailed" : e));
	}

	async function resend(): Promise<void> {
		if (!unverifiedEmail) return;
		const callbackURL = getPathname({ href: "/verify-email", locale });
		if (await run(() => authClient.sendVerificationEmail({ email: unverifiedEmail, callbackURL }))) setResent(true);
	}

	const errorBox = error && (
		<Notice tone="error">
			{t(`errors.${error}`)}
			{unverifiedEmail && !resent && (
				<button
					type="button"
					onClick={resend}
					className="ml-1 font-medium text-fg underline underline-offset-4"
				>
					{t("login.resend")}
				</button>
			)}
		</Notice>
	);

	if (step !== "credentials") {
		return (
			<form onSubmit={onCode} className="flex flex-col gap-4" noValidate>
				<h2 className="text-lg font-semibold">{t("login.twoFactorTitle")}</h2>
				<p className="text-sm text-muted">
					{step === "totp" ? t("login.twoFactorHelp") : t("login.backupHelp")}
				</p>
				{errorBox}
				<Field id="code" label={t("login.code")}>
					<Input
						id="code"
						name="code"
						required
						autoFocus
						autoComplete="one-time-code"
						inputMode={step === "totp" ? "numeric" : "text"}
						pattern={step === "totp" ? "[0-9]{6}" : undefined}
						maxLength={step === "totp" ? 6 : 32}
						className="font-mono tracking-widest"
					/>
				</Field>
				<label className="flex min-h-11 items-center gap-2 text-sm">
					<input type="checkbox" name="trustDevice" className="size-5" />
					{t("login.trustDevice")}
				</label>
				<button type="submit" disabled={pending} className={buttonClass("primary")}>
					{t("login.verify")}
				</button>
				<button
					type="button"
					onClick={() => {
						setError(null);
						setStep(step === "totp" ? "backup" : "totp");
					}}
					className="min-h-11 text-sm underline underline-offset-4"
				>
					{step === "totp" ? t("login.useBackup") : t("login.useTotp")}
				</button>
			</form>
		);
	}

	return (
		<div className="flex flex-col gap-4">
			<form onSubmit={onCredentials} className="flex flex-col gap-4" noValidate>
				{errorBox}
				{resent && <Notice tone="success">{t("login.resent")}</Notice>}
				<Field id="identifier" label={t("fields.identifier")}>
					<Input id="identifier" name="identifier" required autoComplete="username webauthn" />
				</Field>
				<Field id="password" label={t("fields.password")}>
					<PasswordInput id="password" name="password" required autoComplete="current-password" />
				</Field>
				<Link href="/recover-password" className="self-end text-sm underline underline-offset-4">
					{t("login.forgot")}
				</Link>
				<button type="submit" disabled={pending} className={buttonClass("primary")}>
					{t("login.submit")}
				</button>
			</form>
			<Divider label={t("or")} />
			{googleEnabled && <GoogleButton callbackURL={next} />}
			<button type="button" onClick={onPasskey} disabled={pending} className={buttonClass("secondary", "w-full")}>
				<svg
					aria-hidden="true"
					viewBox="0 0 24 24"
					className="size-5"
					fill="none"
					stroke="currentColor"
					strokeWidth="2"
				>
					<circle cx="8" cy="8" r="4" />
					<path d="M2 21v-2a6 6 0 0 1 9-5.2M15 11a3 3 0 1 1 3 3v7l-2-1.5L18 18l-2-1.5 2-1.5" />
				</svg>
				{t("passkey")}
			</button>
		</div>
	);
}
