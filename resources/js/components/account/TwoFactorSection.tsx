import { router, useForm } from "@inertiajs/react";
import type { FormEvent } from "react";

import { useLocalized } from "../../lib/useLocalized";
import { useT } from "../../lib/useT";
import Field from "../Field";
import PasswordField from "../PasswordField";

/** The `twoFactor` prop of Profile/Index (account.TwoFactorProps in app/controllers/account/two_factor.go). */
export interface TwoFactorState {
	enabled: boolean;
	pending?: boolean;
	secret?: string | null;
	qr?: string | null;
	recoveryCodes?: string[] | null;
	recoveryCodesLeft?: number;
}

interface Props {
	state: TwoFactorState;
	button: string;
}

// The posts land back on /profile; preserveScroll keeps the reader at this section.
const stay = { preserveScroll: true } as const;

// TwoFactorSection is the /profile two-factor settings (formerly /user/two-factor): enrol with a QR code or the
// setup key, confirm with a first code, see the recovery codes once, regenerate them or turn 2FA off. The last
// two ask for the current password; the server rate-limits every post.
export default function TwoFactorSection({ state, button }: Props): React.ReactElement {
	const t = useT();
	const l = useLocalized();
	const confirm = useForm({ code: "" });
	const regenerate = useForm({ password: "" });
	const disable = useForm({ password: "" });

	const submitConfirm = (e: FormEvent) => {
		e.preventDefault();
		confirm.post(l("/user/two-factor/confirm"), {
			...stay,
			onFinish: () => confirm.reset("code"),
		});
	};
	const submitRegenerate = (e: FormEvent) => {
		e.preventDefault();
		regenerate.post(l("/user/two-factor/recovery-codes"), {
			...stay,
			onFinish: () => regenerate.reset("password"),
		});
	};
	const submitDisable = (e: FormEvent) => {
		e.preventDefault();
		disable.post(l("/user/two-factor/disable"), {
			...stay,
			onFinish: () => disable.reset("password"),
		});
	};

	if (state.enabled) {
		return (
			<div className="space-y-4">
				<p>{t("account.twoFactor.on")}</p>
				{state.recoveryCodes ? (
					<div role="status">
						<h3 className="font-semibold">{t("account.twoFactor.backupTitle")}</h3>
						<p>{t("account.twoFactor.backupBody")}</p>
						<ul
							className="mt-2 grid grid-cols-2 gap-2 font-mono"
							data-testid="recovery-codes"
						>
							{state.recoveryCodes.map((code) => (
								<li key={code}>{code}</li>
							))}
						</ul>
					</div>
				) : (
					<p>
						{t("account.twoFactor.recoveryLeft", {
							count: state.recoveryCodesLeft ?? 0,
						})}
					</p>
				)}
				<form onSubmit={submitRegenerate} className="space-y-4">
					<PasswordField
						label={t("account.twoFactor.currentPassword")}
						name="password"
						autoComplete="current-password"
						value={regenerate.data.password}
						onChange={(value) => regenerate.setData("password", value)}
						error={regenerate.errors.password}
					/>
					<button type="submit" disabled={regenerate.processing} className={button}>
						{t("account.twoFactor.regenerate")}
					</button>
				</form>
				<h3 className="pt-4 font-semibold">{t("account.twoFactor.turnOff")}</h3>
				<form onSubmit={submitDisable} className="space-y-4">
					<PasswordField
						label={t("account.twoFactor.confirmPassword")}
						name="password"
						autoComplete="current-password"
						value={disable.data.password}
						onChange={(value) => disable.setData("password", value)}
						error={disable.errors.password}
					/>
					<button type="submit" disabled={disable.processing} className={button}>
						{t("account.twoFactor.disable")}
					</button>
				</form>
			</div>
		);
	}

	if (state.pending && state.qr) {
		return (
			<div className="space-y-4">
				<p>{t("account.twoFactor.scan")}</p>
				<img
					className="w-fit"
					src={state.qr}
					width={200}
					height={200}
					alt={t("account.twoFactor.qrAlt")}
				/>
				<p>
					{t("account.twoFactor.manual")}{" "}
					<code className="font-mono select-all">{state.secret}</code>
				</p>
				<form onSubmit={submitConfirm} className="space-y-4">
					<Field
						label={t("account.twoFactor.code")}
						name="code"
						inputMode="numeric"
						autoComplete="one-time-code"
						required
						value={confirm.data.code}
						onChange={(e) => confirm.setData("code", e.target.value)}
						error={confirm.errors.code}
					/>
					<button type="submit" disabled={confirm.processing} className={button}>
						{t("account.twoFactor.verify")}
					</button>
				</form>
			</div>
		);
	}

	return (
		<div className="space-y-4">
			<p>{t("account.twoFactor.off")}</p>
			<button
				type="button"
				className={button}
				onClick={() => router.post(l("/user/two-factor"), {}, stay)}
			>
				{t("account.twoFactor.enable")}
			</button>
		</div>
	);
}
