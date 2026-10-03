"use client";

import { useRouter } from "next/navigation";
import { useTranslations } from "next-intl";
import { type FormEvent, useState } from "react";
import { PasswordInput } from "@/components/auth/password-input";
import { buttonClass, Field, Input, Notice } from "@/components/ui/form";
import { authClient } from "@/lib/auth-client";
import { type AuthClientError, authErrorKey } from "@/lib/auth-errors";

type Props = { enabled: boolean; needsPassword: boolean };
type Setup = { qr: string; secret: string; backupCodes: string[] };
type Step = "idle" | "password" | "scan" | "backup";

export function TwoFactorSection({ enabled, needsPassword }: Props) {
	const t = useTranslations("account.twoFactor");
	const tAuth = useTranslations("auth");
	const router = useRouter();
	const [step, setStep] = useState<Step>("idle");
	const [setup, setSetup] = useState<Setup | null>(null);
	const [error, setError] = useState<string | null>(null);
	const [notice, setNotice] = useState<string | null>(null);
	const [copied, setCopied] = useState(false);
	const [pending, setPending] = useState(false);

	async function call<T>(fn: () => Promise<{ data: T | null; error: AuthClientError }>): Promise<T | null> {
		setPending(true);
		setError(null);
		try {
			const { data, error } = await fn();
			if (error) setError(tAuth(`errors.${authErrorKey(error)}`));
			return error ? null : data;
		} catch {
			setError(tAuth("errors.network"));
			return null;
		} finally {
			setPending(false);
		}
	}

	async function onPassword(event: FormEvent<HTMLFormElement>): Promise<void> {
		event.preventDefault();
		await toggle(String(new FormData(event.currentTarget).get("password") ?? ""));
	}

	async function toggle(password: string | undefined): Promise<void> {
		if (enabled) {
			if (await call(() => authClient.twoFactor.disable({ password }))) {
				setStep("idle");
				setNotice(t("disabled"));
				router.refresh();
			}
			return;
		}
		const data = await call(() => authClient.twoFactor.enable({ password, method: "totp" }));
		if (data?.method !== "totp") return;
		// Loaded on demand: only people enabling 2FA download the QR encoder.
		const { toDataURL } = await import("qrcode");
		setSetup({
			qr: await toDataURL(data.totpURI, { margin: 1, width: 208 }),
			secret: new URL(data.totpURI).searchParams.get("secret") ?? "",
			backupCodes: data.backupCodes,
		});
		setStep("scan");
	}

	async function onVerify(event: FormEvent<HTMLFormElement>): Promise<void> {
		event.preventDefault();
		const code = String(new FormData(event.currentTarget).get("code") ?? "").trim();
		if (await call(() => authClient.twoFactor.verifyTotp({ code }))) setStep("backup");
	}

	async function start(): Promise<void> {
		setNotice(null);
		// Passwordless accounts (Google/passkey only) skip the confirmation step.
		if (needsPassword) setStep("password");
		else await toggle(undefined);
	}

	const errorBox = error && <Notice tone="error">{error}</Notice>;

	if (step === "password") {
		return (
			<form method="post" onSubmit={onPassword} className="flex flex-col gap-4" noValidate>
				{errorBox}
				<Field id="tf-password" label={t("confirmPassword")}>
					<PasswordInput
						id="tf-password"
						name="password"
						required
						autoComplete="current-password"
						autoFocus
					/>
				</Field>
				<div className="flex flex-wrap gap-2">
					<button type="submit" disabled={pending} className={buttonClass(enabled ? "danger" : "primary")}>
						{enabled ? t("disable") : t("enable")}
					</button>
					<button type="button" onClick={() => setStep("idle")} className={buttonClass("secondary")}>
						{t("cancel")}
					</button>
				</div>
			</form>
		);
	}

	if (step === "scan" && setup) {
		return (
			<form method="post" onSubmit={onVerify} className="flex flex-col gap-4" noValidate>
				<p className="text-sm">{t("scan")}</p>
				{/* biome-ignore lint/performance/noImgElement: data: URL generated in the browser, nothing to optimize */}
				<img src={setup.qr} alt={t("qrAlt")} width={208} height={208} className="rounded-md bg-white" />
				<p className="text-sm text-muted">
					{t("manual")} <code className="break-all font-mono text-fg">{setup.secret}</code>
				</p>
				{errorBox}
				<Field id="tf-code" label={t("code")}>
					<Input
						id="tf-code"
						name="code"
						required
						inputMode="numeric"
						autoComplete="one-time-code"
						pattern="[0-9]{6}"
						maxLength={6}
						className="font-mono tracking-widest"
					/>
				</Field>
				<button type="submit" disabled={pending} className={buttonClass("primary", "self-start")}>
					{t("verify")}
				</button>
			</form>
		);
	}

	if (step === "backup" && setup) {
		return (
			<div className="flex flex-col gap-4">
				<h3 className="font-semibold">{t("backupTitle")}</h3>
				<p className="text-sm text-muted">{t("backupBody")}</p>
				<ul className="grid grid-cols-2 gap-1 rounded-md bg-surface p-3 font-mono text-sm">
					{setup.backupCodes.map((code) => (
						<li key={code}>{code}</li>
					))}
				</ul>
				<div className="flex flex-wrap gap-2">
					<button
						type="button"
						onClick={async () => {
							await navigator.clipboard.writeText(setup.backupCodes.join("\n"));
							setCopied(true);
						}}
						className={buttonClass("secondary")}
					>
						{copied ? t("copied") : t("copy")}
					</button>
					<button
						type="button"
						onClick={() => {
							setSetup(null);
							setStep("idle");
							router.refresh();
						}}
						className={buttonClass("primary")}
					>
						{t("done")}
					</button>
				</div>
			</div>
		);
	}

	return (
		<div className="flex flex-col gap-4">
			{notice && <Notice tone="success">{notice}</Notice>}
			{errorBox}
			<p className="text-sm">{enabled ? t("on") : t("off")}</p>
			<button
				type="button"
				onClick={start}
				disabled={pending}
				className={buttonClass(enabled ? "secondary" : "primary", "self-start")}
			>
				{enabled ? t("disable") : t("enable")}
			</button>
		</div>
	);
}
