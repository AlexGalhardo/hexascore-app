import { Head, Link, useForm } from "@inertiajs/react";
import { type ClipboardEvent, type FormEvent, type KeyboardEvent, useRef, useState } from "react";

import AuthFrame, { AuthFooter, authLink } from "../../components/auth/AuthFrame";
import Field from "../../components/Field";
import { btnClass } from "../../components/ui/Btn";
import { useLocalized } from "../../lib/useLocalized";
import { useT } from "../../lib/useT";

const DIGITS = 6;
const empty = (): string[] => Array(DIGITS).fill("");

/**
 * The second login step (owner 2026-11-07, its own page after the dialog of 2026-11-02): six one-digit boxes that
 * accept a pasted code, move with the keys and submit on their own once the sixth digit is in. A recovery code is
 * the fallback. The server answers a wrong code with an error here and keeps the challenge.
 */
export default function TwoFactorChallenge() {
	const t = useT();
	const l = useLocalized();
	const [recovery, setRecovery] = useState(false);
	const [digits, setDigits] = useState<string[]>(empty);
	const boxes = useRef<(HTMLInputElement | null)[]>([]);
	const form = useForm({ code: "", recovery_code: "" });

	const send = (code: string, recoveryCode = "") => {
		form.transform(() => ({ code, recovery_code: recoveryCode }));
		form.post(l("/2fa"), {
			preserveScroll: true,
			onError: () => {
				setDigits(empty());
				boxes.current[0]?.focus();
			},
		});
	};

	const fill = (at: number, typed: string) => {
		const clean = typed.replace(/\D/g, "");
		if (!clean) return;
		const next = [...digits];
		for (let i = 0; i < clean.length && at + i < DIGITS; i++) next[at + i] = clean[i];
		setDigits(next);
		boxes.current[Math.min(at + clean.length, DIGITS - 1)]?.focus();
		if (next.every((d) => d !== "") && !form.processing) send(next.join(""));
	};

	const onKeyDown = (at: number, e: KeyboardEvent<HTMLInputElement>) => {
		if (e.key === "Backspace" && !digits[at] && at > 0) {
			e.preventDefault();
			const next = [...digits];
			next[at - 1] = "";
			setDigits(next);
			boxes.current[at - 1]?.focus();
		} else if (e.key === "ArrowLeft" && at > 0) {
			boxes.current[at - 1]?.focus();
		} else if (e.key === "ArrowRight" && at < DIGITS - 1) {
			boxes.current[at + 1]?.focus();
		}
	};

	const onPaste = (at: number, e: ClipboardEvent<HTMLInputElement>) => {
		e.preventDefault();
		fill(at, e.clipboardData.getData("text"));
	};

	const swap = () => {
		// Only one field is sent filled: the server checks the app code first.
		setRecovery(!recovery);
		setDigits(empty());
		form.reset();
		form.clearErrors();
	};

	return (
		<AuthFrame title={t("auth.login.twoFactorTitle")}>
			<Head title={t("auth.login.twoFactorTitle")} />
			<p className="-mt-4 mb-8 text-center text-muted">
				{recovery ? t("auth.login.backupHelp") : t("auth.login.twoFactorHelp")}
			</p>
			{recovery ? (
				<form
					method="post"
					action={l("/2fa")}
					className="space-y-5"
					onSubmit={(e: FormEvent) => {
						e.preventDefault();
						send("", form.data.recovery_code);
					}}
				>
					<Field
						label={t("auth.login.backupCode")}
						name="recovery_code"
						autoComplete="off"
						autoFocus
						required
						value={form.data.recovery_code}
						onChange={(e) => form.setData("recovery_code", e.target.value)}
						error={form.errors.recovery_code}
					/>
					<button
						type="submit"
						disabled={form.processing}
						className={btnClass("primary", "w-full")}
					>
						{t("auth.login.verify")}
					</button>
				</form>
			) : (
				<form
					method="post"
					action={l("/2fa")}
					onSubmit={(e: FormEvent) => {
						e.preventDefault();
						if (digits.every((d) => d !== "")) send(digits.join(""));
					}}
				>
					<fieldset disabled={form.processing}>
						<legend className="sr-only">{t("auth.login.code")}</legend>
						<div className="flex justify-between gap-2">
							{digits.map((d, i) => (
								<input
									// biome-ignore lint/suspicious/noArrayIndexKey: the six boxes are fixed positions
									key={i}
									ref={(el) => {
										boxes.current[i] = el;
									}}
									aria-label={`${t("auth.login.code")} ${i + 1}`}
									inputMode="numeric"
									pattern="[0-9]*"
									autoComplete={i === 0 ? "one-time-code" : "off"}
									// biome-ignore lint/a11y/noAutofocus: the page has one task, the code
									autoFocus={i === 0}
									maxLength={DIGITS}
									value={d}
									onChange={(e) => fill(i, e.target.value.slice(-DIGITS))}
									onKeyDown={(e) => onKeyDown(i, e)}
									onPaste={(e) => onPaste(i, e)}
									className="size-12 rounded-md border-2 border-field bg-field-bg text-center text-2xl font-bold tabular-nums text-fg shadow-ledge-secondary focus-visible:border-primary focus-visible:outline-none"
								/>
							))}
						</div>
						{form.errors.code && (
							<p role="alert" className="mt-3 text-center text-sm text-danger">
								{form.errors.code}
							</p>
						)}
					</fieldset>
					<noscript>
						<Field
							label={t("auth.login.code")}
							name="code"
							inputMode="numeric"
							autoComplete="one-time-code"
						/>
						<button type="submit" className={btnClass("primary", "mt-5 w-full")}>
							{t("auth.login.verify")}
						</button>
					</noscript>
				</form>
			)}
			<AuthFooter>
				<button type="button" onClick={swap} className={authLink}>
					{recovery ? t("auth.login.useTotp") : t("auth.login.useBackup")}
				</button>
				{" · "}
				<Link href={l("/login")} className={authLink}>
					{t("auth.login.title")}
				</Link>
			</AuthFooter>
		</AuthFrame>
	);
}
