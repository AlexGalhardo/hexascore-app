import { Head, Link, useForm } from "@inertiajs/react";
import { type FormEvent, useState } from "react";

import ChallengeStatus from "../../components/ChallengeStatus";
import Field from "../../components/Field";
import PasswordField from "../../components/PasswordField";
import SubmitButton from "../../components/SubmitButton";
import { challengeHeaders } from "../../lib/challenge";
import { useLocalized } from "../../lib/useLocalized";
import { useT } from "../../lib/useT";
import { getPasskey, PasskeyCancelled } from "../../lib/webauthn";

export function PasskeySignIn(): React.ReactElement {
	const t = useT();
	const l = useLocalized();
	const [failed, setFailed] = useState(false);

	const signIn = async (): Promise<void> => {
		setFailed(false);

		try {
			const options = await fetch(l("/login/passkey/options"), {
				method: "POST",
				credentials: "same-origin",
				headers: { Accept: "application/json", "Content-Type": "application/json" },
				body: "{}",
			});
			if (!options.ok) {
				throw new PasskeyCancelled();
			}

			const credential = await getPasskey(await options.json());
			const res = await fetch(l("/login/passkey"), {
				method: "POST",
				credentials: "same-origin",
				headers: { Accept: "application/json", "Content-Type": "application/json" },
				body: JSON.stringify(credential),
			});
			if (!res.ok) {
				throw new PasskeyCancelled();
			}

			window.location.assign(res.url);
		} catch (err) {
			if (err instanceof PasskeyCancelled) {
				setFailed(true);
				return;
			}

			setFailed(true);
		}
	};

	return (
		<div className="mt-4">
			<button
				type="button"
				onClick={() => void signIn()}
				className="rounded-md border border-zinc-300 px-4 py-2 font-medium hover:bg-zinc-50 dark:border-zinc-700 dark:hover:bg-zinc-900"
			>
				{t("auth.passkey")}
			</button>
			{failed && (
				<p role="alert" className="mt-2 text-sm text-red-600">
					{t("auth.errors.passkeyFailed")}
				</p>
			)}
		</div>
	);
}

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
			<form method="post" action={l("/login")} onSubmit={submit} className="space-y-4">
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
				<SubmitButton pending={form.processing || checking}>
					{t("auth.login.submit")}
				</SubmitButton>
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
			<PasskeySignIn />
		</>
	);
}
