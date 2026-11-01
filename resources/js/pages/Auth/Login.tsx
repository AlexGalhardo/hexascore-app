import { Head, Link, useForm } from "@inertiajs/react";
import { type FormEvent, useState } from "react";

import ChallengeStatus from "../../components/ChallengeStatus";
import Field from "../../components/Field";
import PasswordField from "../../components/PasswordField";
import { challengeHeaders } from "../../lib/challenge";
import { useLocalized } from "../../lib/useLocalized";
import { useT } from "../../lib/useT";

export default function Login() {
	const t = useT();
	const l = useLocalized();
	const form = useForm({ identifier: "", password: "", remember: false });
	const [checking, setChecking] = useState(false);

	const submit = async (e: FormEvent) => {
		e.preventDefault();
		const headers = await challengeHeaders("login", setChecking);
		form.post(l("/login"), { headers, onFinish: () => form.reset("password") });
	};

	return (
		<>
			<Head title={t("auth.login.metaTitle")} />
			<h1 className="mb-4 text-2xl font-semibold">{t("auth.login.title")}</h1>
			<form onSubmit={submit} className="space-y-4">
				<Field
					label={t("auth.fields.identifier")}
					autoComplete="username"
					required
					value={form.data.identifier}
					onChange={(e) => form.setData("identifier", e.target.value)}
					error={form.errors.identifier}
				/>
				{/* No rule checklist here: logging in must not advertise the policy,
            and an existing password may predate the current rules. */}
				<PasswordField
					label={t("auth.fields.password")}
					name="password"
					autoComplete="current-password"
					value={form.data.password}
					onChange={(value) => form.setData("password", value)}
					error={form.errors.password}
				/>
				<label className="flex items-center gap-2">
					<input
						type="checkbox"
						checked={form.data.remember}
						onChange={(e) => form.setData("remember", e.target.checked)}
					/>
					{t("auth.login.rememberMe")}
				</label>
				<button
					type="submit"
					disabled={form.processing}
					className="rounded-md bg-indigo-600 px-4 py-2 font-medium text-white hover:bg-indigo-500 disabled:opacity-50"
				>
					{t("auth.login.submit")}
				</button>
				<ChallengeStatus checking={checking} />
			</form>
			<p className="mt-6 flex gap-4 text-sm">
				<Link
					href="/recover-password"
					className="text-indigo-600 hover:underline dark:text-indigo-400"
				>
					{t("auth.login.forgot")}
				</Link>
				<Link
					href="/signup"
					className="text-indigo-600 hover:underline dark:text-indigo-400"
				>
					{t("auth.login.createAccount")}
				</Link>
			</p>
		</>
	);
}
