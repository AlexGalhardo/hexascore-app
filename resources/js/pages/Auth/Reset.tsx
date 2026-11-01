import { Head, useForm } from "@inertiajs/react";
import type { FormEvent } from "react";
import AuthFrame from "../../components/auth/AuthFrame";
import Field from "../../components/Field";
import PasswordField from "../../components/PasswordField";
import SubmitButton from "../../components/SubmitButton";
import { checkConfirmation, checkPassword, passwordAdvice } from "../../lib/password-policy";
import { useLocalized } from "../../lib/useLocalized";
import { useT } from "../../lib/useT";

export default function Reset({ token, email }: { token: string; email: string }) {
	const t = useT();
	const l = useLocalized();
	const form = useForm({ token, email, password: "", password_confirmation: "" });

	const submit = (e: FormEvent) => {
		e.preventDefault();
		form.post(l("/reset-password"), {
			onFinish: () => form.reset("password", "password_confirmation"),
		});
	};

	return (
		<AuthFrame heading={t("auth.recover.heading")} title={t("auth.reset.title")}>
			<Head title={t("auth.reset.metaTitle")} />
			<form onSubmit={submit} className="space-y-5">
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
					label={t("auth.fields.newPassword")}
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
				<SubmitButton pending={form.processing} className="w-full">
					{t("auth.reset.submit")}
				</SubmitButton>
			</form>
		</AuthFrame>
	);
}
