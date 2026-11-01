import { Head, useForm } from "@inertiajs/react";
import type { FormEvent } from "react";

import Field from "../../components/Field";
import {
	NICKNAME_ERRORS,
	NICKNAME_MAX,
	NICKNAME_MIN,
	NICKNAME_PATTERN,
	normalizeNickname,
} from "../../lib/signup-policy";
import { useLocalized } from "../../lib/useLocalized";
import { useT } from "../../lib/useT";

export default function Onboarding({ socialEnabled = false }: { socialEnabled?: boolean }) {
	const t = useT();
	const l = useLocalized();
	const form = useForm({ nickname: "", feedOptIn: false });

	const submit = (e: FormEvent) => {
		e.preventDefault();
		form.post(l("/onboarding"));
	};

	const code = form.errors.nickname;
	const error = code ? t(NICKNAME_ERRORS[code] ?? "auth.errors.generic") : undefined;

	return (
		<>
			<Head title={t("auth.onboarding.metaTitle")} />
			<h1 className="mb-2 text-2xl font-semibold">{t("auth.onboarding.title")}</h1>
			<p className="mb-4 text-zinc-600 dark:text-zinc-400">{t("auth.onboarding.body")}</p>
			<form onSubmit={submit} className="space-y-4">
				<Field
					label={t("auth.fields.nickname")}
					autoComplete="username"
					autoCapitalize="none"
					minLength={NICKNAME_MIN}
					maxLength={NICKNAME_MAX}
					pattern={NICKNAME_PATTERN}
					required
					value={form.data.nickname}
					onChange={(e) => form.setData("nickname", normalizeNickname(e.target.value))}
					error={error}
				/>
				<p className="text-sm text-zinc-600 dark:text-zinc-400">
					{t("auth.signup.nicknameHint")}
				</p>
				{socialEnabled && (
					<label className="flex items-start gap-2">
						<input
							type="checkbox"
							checked={form.data.feedOptIn}
							onChange={(e) => form.setData("feedOptIn", e.target.checked)}
						/>
						<span>
							<span className="font-medium">{t("social.optIn.label")}</span>
							<span className="block text-sm text-zinc-600 dark:text-zinc-400">
								{t("social.optIn.help")}
							</span>
						</span>
					</label>
				)}
				<button
					type="submit"
					disabled={form.processing}
					className="rounded-md bg-primary px-4 py-2 font-medium text-on-primary hover:bg-primary-hover disabled:opacity-50"
				>
					{t("auth.onboarding.submit")}
				</button>
			</form>
		</>
	);
}
