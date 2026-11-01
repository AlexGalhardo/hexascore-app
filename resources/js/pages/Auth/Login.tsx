import { Head, Link, useForm } from "@inertiajs/react";
import { type FormEvent, useState } from "react";

import AuthFrame, { authLink, GoogleButton, OrRule } from "../../components/auth/AuthFrame";
import ChallengeStatus from "../../components/ChallengeStatus";
import Field from "../../components/Field";
import PasswordField from "../../components/PasswordField";
import SubmitButton from "../../components/SubmitButton";
import { btnClass } from "../../components/ui/Btn";
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
				className={btnClass("secondary", "w-full")}
			>
				{t("auth.passkey")}
			</button>
			{failed && (
				<p role="alert" className="mt-2 text-sm text-danger">
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
		<AuthFrame
			heading={t("auth.login.heading")}
			lede={t("auth.login.lede")}
			title={t("auth.login.title")}
		>
			<Head title={t("auth.login.metaTitle")} />
			<GoogleButton mode="signIn" />
			<OrRule />
			<form method="post" action={l("/login")} onSubmit={submit} className="space-y-5">
				<Field
					label={t("auth.fields.identifier")}
					name="identifier"
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
				<div className="flex flex-wrap items-center justify-between gap-x-4 gap-y-2 pt-1">
					<label className="flex min-h-11 items-center gap-2.5">
						<input
							type="checkbox"
							name="remember"
							value="true"
							checked={form.data.remember}
							onChange={(e) => form.setData("remember", e.target.checked)}
							className="size-5 accent-fg"
						/>
						{t("auth.login.rememberMe")}
					</label>
					<Link href={l("/recover-password")} className={`${authLink} text-sm`}>
						{t("auth.login.forgot")}
					</Link>
				</div>
				<SubmitButton pending={form.processing || checking} className="w-full">
					{t("auth.login.submit")}
				</SubmitButton>
				<ChallengeStatus checking={checking} />
			</form>
			<PasskeySignIn />
			<p className="mt-8 text-muted">
				{t("auth.login.newHere")}{" "}
				<Link href={l("/signup")} className={authLink}>
					{t("auth.login.createAccount")}
				</Link>
			</p>
		</AuthFrame>
	);
}
