import { Head, Link, useForm } from "@inertiajs/react";
import { type FormEvent, useState } from "react";

import ChallengeStatus from "../../components/ChallengeStatus";
import Field from "../../components/Field";
import PasswordField from "../../components/PasswordField";
import SubmitButton from "../../components/SubmitButton";
import { challengeHeaders } from "../../lib/challenge";
import { checkConfirmation, checkPassword, passwordAdvice } from "../../lib/password-policy";
import { useLocalized } from "../../lib/useLocalized";
import { useT } from "../../lib/useT";

export default function Signup({ formToken }: { formToken: string }) {
	const t = useT();
	const l = useLocalized();
	const form = useForm({
		name: "",
		email: "",
		password: "",
		password_confirmation: "",
		website: "",
		form_token: formToken,
	});
	const [checking, setChecking] = useState(false);

	const submit = async (e: FormEvent) => {
		e.preventDefault();
		const headers = await challengeHeaders("signup", setChecking);
		form.post(l("/signup"), {
			headers,
			onFinish: () => form.reset("password", "password_confirmation"),
		});
	};

	return (
		<>
			<Head title={t("auth.signup.metaTitle")} />
			<h1 className="mb-4 text-2xl font-semibold">{t("auth.signup.title")}</h1>
			<form method="post" action={l("/signup")} onSubmit={submit} className="space-y-4">
				{/* Honeypot: invisible to people and assistive tech, bots fill it. */}
				<input
					type="text"
					name="website"
					tabIndex={-1}
					autoComplete="off"
					aria-hidden="true"
					className="absolute -left-[9999px]"
					value={form.data.website}
					onChange={(e) => form.setData("website", e.target.value)}
				/>
				<input type="hidden" name="form_token" value={form.data.form_token} />
				<Field
					label={t("auth.fields.name")}
					name="name"
					autoComplete="name"
					required
					value={form.data.name}
					onChange={(e) => form.setData("name", e.target.value)}
					error={form.errors.name}
				/>
				<Field
					label={t("auth.fields.email")}
					name="email"
					type="email"
					autoComplete="email"
					required
					value={form.data.email}
					onChange={(e) => form.setData("email", e.target.value)}
					error={form.errors.email}
				/>
				<PasswordField
					label={t("auth.fields.password")}
					name="password"
					autoComplete="new-password"
					value={form.data.password}
					onChange={(value) => form.setData("password", value)}
					error={form.errors.password}
					checks={checkPassword(form.data.password)}
					advice={passwordAdvice(form.data.password)}
				/>
				<PasswordField
					label={t("auth.fields.confirmPassword")}
					name="password_confirmation"
					autoComplete="new-password"
					value={form.data.password_confirmation}
					onChange={(value) => form.setData("password_confirmation", value)}
					error={form.errors.password_confirmation}
					checks={checkConfirmation(form.data.password, form.data.password_confirmation)}
				/>
				<SubmitButton pending={form.processing || checking}>
					{t("auth.signup.submit")}
				</SubmitButton>
				<ChallengeStatus checking={checking} />
			</form>
			<p className="mt-6 text-sm">
				<Link
					href="/login"
					className="text-indigo-600 hover:underline dark:text-indigo-400"
				>
					{t("auth.signup.haveAccount")} {t("auth.signup.logIn")}
				</Link>
			</p>
		</>
	);
}
