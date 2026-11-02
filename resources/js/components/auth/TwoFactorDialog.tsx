import { useForm } from "@inertiajs/react";
import { type ClipboardEvent, type KeyboardEvent, useRef, useState } from "react";

import { useLocalized } from "../../lib/useLocalized";
import { useT } from "../../lib/useT";
import Dialog from "../Dialog";
import Field from "../Field";
import { btnClass } from "../ui/Btn";

const DIGITS = 6;

/**
 * The second login step as a dialog over the login page (owner 2026-11-02): six one-digit boxes that accept a
 * pasted code, move with the keys, and submit on their own once the sixth digit is in. A recovery code is the
 * fallback. The dialog stays open while the login waits for its code; the server answers a wrong code with an error
 * here and keeps the challenge.
 */
export default function TwoFactorDialog({ open }: { open: boolean }) {
	const t = useT();
	const l = useLocalized();
	const [recovery, setRecovery] = useState(false);
	const [digits, setDigits] = useState<string[]>(Array(DIGITS).fill(""));
	const boxes = useRef<(HTMLInputElement | null)[]>([]);
	const form = useForm({ code: "", recovery_code: "" });

	const send = (code: string, recoveryCode = "") => {
		form.transform(() => ({ code, recovery_code: recoveryCode }));
		form.post(l("/two-factor-challenge"), {
			preserveScroll: true,
			onError: () => {
				setDigits(Array(DIGITS).fill(""));
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

	return (
		// Not dismissible: the login is half done until a code is checked; signing in again starts over.
		<Dialog
			open={open}
			onOpenChange={() => {}}
			title={t("auth.login.twoFactorTitle")}
			description={recovery ? t("auth.login.backupHelp") : t("auth.login.twoFactorHelp")}
			size="sm"
		>
			{recovery ? (
				<form
					className="mt-4 space-y-4"
					onSubmit={(e) => {
						e.preventDefault();
						send("", form.data.recovery_code);
					}}
				>
					<Field
						label={t("auth.login.backupCode")}
						autoComplete="off"
						required
						value={form.data.recovery_code}
						onChange={(e) => form.setData("recovery_code", e.target.value)}
						error={form.errors.recovery_code}
					/>
					<button
						type="submit"
						disabled={form.processing}
						className={btnClass("primary")}
					>
						{t("auth.login.verify")}
					</button>
				</form>
			) : (
				<fieldset className="mt-4" disabled={form.processing}>
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
								autoComplete={i === 0 ? "one-time-code" : "off"}
								// biome-ignore lint/a11y/noAutofocus: the dialog exists to take this code
								autoFocus={i === 0}
								maxLength={DIGITS}
								value={d}
								onChange={(e) => fill(i, e.target.value.slice(-DIGITS))}
								onKeyDown={(e) => onKeyDown(i, e)}
								onPaste={(e) => onPaste(i, e)}
								className="size-12 rounded-xs border-2 border-field bg-field-bg text-center text-2xl font-bold tabular-nums text-fg shadow-ledge-secondary focus-visible:border-primary focus-visible:outline-none"
							/>
						))}
					</div>
					{form.errors.code && (
						<p role="alert" className="mt-3 text-sm text-danger">
							{form.errors.code}
						</p>
					)}
				</fieldset>
			)}
			<p className="mt-6 text-sm">
				<button
					type="button"
					onClick={() => {
						// Only one field is sent filled: the server checks the app code first.
						setRecovery(!recovery);
						setDigits(Array(DIGITS).fill(""));
						form.reset();
						form.clearErrors();
					}}
					className="text-primary hover:underline"
				>
					{recovery ? t("auth.login.useTotp") : t("auth.login.useBackup")}
				</button>
			</p>
		</Dialog>
	);
}
