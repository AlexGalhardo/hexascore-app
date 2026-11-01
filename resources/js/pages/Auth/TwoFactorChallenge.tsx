import { Head, useForm } from "@inertiajs/react";
import { type FormEvent, useState } from "react";

import Field from "../../components/Field";
import { useLocalized } from "../../lib/useLocalized";
import { useT } from "../../lib/useT";

export default function TwoFactorChallenge() {
	const t = useT();
	const l = useLocalized();
	const [recovery, setRecovery] = useState(false);
	const form = useForm({ code: "", recovery_code: "" });

	const submit = (e: FormEvent) => {
		e.preventDefault();
		form.post(l("/two-factor-challenge"), { onFinish: () => form.reset() });
	};

	return (
		<>
			<Head title={t("auth.login.twoFactorTitle")} />
			<h1 className="mb-4 text-2xl font-semibold">{t("auth.login.twoFactorTitle")}</h1>
			<p className="mb-4">
				{recovery ? t("auth.login.backupHelp") : t("auth.login.twoFactorHelp")}
			</p>
			<form onSubmit={submit} className="space-y-4">
				{recovery ? (
					<Field
						label={t("auth.login.backupCode")}
						autoComplete="off"
						required
						value={form.data.recovery_code}
						onChange={(e) => form.setData("recovery_code", e.target.value)}
						error={form.errors.recovery_code}
					/>
				) : (
					<Field
						label={t("auth.login.code")}
						inputMode="numeric"
						autoComplete="one-time-code"
						autoFocus
						required
						value={form.data.code}
						onChange={(e) => form.setData("code", e.target.value)}
						error={form.errors.code}
					/>
				)}
				<button
					type="submit"
					disabled={form.processing}
					className="rounded-md bg-primary px-4 py-2 font-medium text-on-primary hover:bg-primary-hover disabled:opacity-50"
				>
					{t("auth.login.verify")}
				</button>
			</form>
			<p className="mt-6 text-sm">
				<button
					type="button"
					onClick={() => {
						// Only one field is sent filled: the server checks the app code first.
						setRecovery(!recovery);
						form.reset();
						form.clearErrors();
					}}
					className="text-primary hover:underline"
				>
					{recovery ? t("auth.login.useTotp") : t("auth.login.useBackup")}
				</button>
			</p>
		</>
	);
}
