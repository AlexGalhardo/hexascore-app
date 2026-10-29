import { Head, Link, useForm } from "@inertiajs/react";
import type { FormEvent } from "react";

import Field from "../../components/Field";
import PasswordField from "../../components/PasswordField";
import { checkConfirmation, checkPassword, passwordAdvice } from "../../lib/password-policy";
import { useT } from "../../lib/useT";

export default function Signup() {
	const t = useT();
	const form = useForm({ name: "", email: "", password: "", password_confirmation: "" });

	const submit = (e: FormEvent) => {
		e.preventDefault();
		form.post("/signup", { onFinish: () => form.reset("password", "password_confirmation") });
	};

	return (
		<>
			<Head title={t("auth.signup.metaTitle")} />
			<h1 className="mb-4 text-2xl font-semibold">{t("auth.signup.title")}</h1>
			<form onSubmit={submit} className="space-y-4">
				<Field
					label={t("auth.fields.name")}
					autoComplete="name"
					required
					value={form.data.name}
					onChange={(e) => form.setData("name", e.target.value)}
					error={form.errors.name}
				/>
				<Field
					label={t("auth.fields.email")}
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
				<button
					type="submit"
					disabled={form.processing}
					className="rounded-md bg-indigo-600 px-4 py-2 font-medium text-white hover:bg-indigo-500 disabled:opacity-50"
				>
					{t("auth.signup.submit")}
				</button>
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
