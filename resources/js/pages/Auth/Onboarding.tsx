import { Head, useForm } from "@inertiajs/react";
import type { FormEvent } from "react";

import Field from "../../components/Field";
import { useLocalized } from "../../lib/useLocalized";
import { useT } from "../../lib/useT";

export default function Onboarding() {
	const t = useT();
	const l = useLocalized();
	const form = useForm({ nickname: "" });

	const submit = (e: FormEvent) => {
		e.preventDefault();
		form.post(l("/onboarding"));
	};

	const codes: Record<string, string> = {
		"nickname-taken": "auth.errors.usernameTaken",
		"invalid-nickname": "auth.errors.invalidNickname",
	};
	const code = form.errors.nickname;
	const error = code ? t(codes[code] ?? "auth.errors.generic") : undefined;

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
					minLength={4}
					maxLength={16}
					pattern="[a-z]{4,16}"
					required
					value={form.data.nickname}
					onChange={(e) => form.setData("nickname", e.target.value.toLowerCase())}
					error={error}
				/>
				<p className="text-sm text-zinc-600 dark:text-zinc-400">
					{t("auth.signup.nicknameHint")}
				</p>
				<button
					type="submit"
					disabled={form.processing}
					className="rounded-md bg-indigo-600 px-4 py-2 font-medium text-white hover:bg-indigo-500 disabled:opacity-50"
				>
					{t("auth.onboarding.submit")}
				</button>
			</form>
		</>
	);
}
