import { Head, Link, useForm } from "@inertiajs/react";
import { type FormEvent, useState } from "react";

import AuthFrame, { authLink, GoogleButton, OrRule } from "../../components/auth/AuthFrame";
import ChallengeStatus from "../../components/ChallengeStatus";
import Field from "../../components/Field";
import PasswordField from "../../components/PasswordField";
import SubmitButton from "../../components/SubmitButton";
import { challengeHeaders } from "../../lib/challenge";
import { checkConfirmation, checkPassword, passwordAdvice } from "../../lib/password-policy";
import { useLocalized } from "../../lib/useLocalized";
import { useT } from "../../lib/useT";

// Renders "...<terms>Terms</terms>...<privacy>Privacy policy</privacy>." with the two tags as links.
export function Agreement({ text, l }: { text: string; l: (path: string) => string }): React.ReactElement {
	const href = { terms: "/terms", privacy: "/policy" } as const;

	return (
		<p className="text-sm text-muted">
			{text.split(/(<(?:terms|privacy)>.*?<\/(?:terms|privacy)>)/).map((part) => {
				const m = /^<(terms|privacy)>(.*)<\/\1>$/.exec(part);
				return m ? (
					<Link key={m[1]} href={l(href[m[1] as keyof typeof href])} className={authLink}>
						{m[2]}
					</Link>
				) : (
					part
				);
			})}
		</p>
	);
}

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
		<AuthFrame
			heading={t("auth.signup.heading")}
			lede={t("auth.signup.lede")}
			title={t("auth.signup.cardTitle")}
		>
			<Head title={t("auth.signup.metaTitle")} />
			<GoogleButton mode="signUp" />
			<OrRule />
			<form method="post" action={l("/signup")} onSubmit={submit} className="space-y-5">
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
				<Agreement text={t("auth.signup.agree")} l={l} />
				<SubmitButton pending={form.processing || checking} className="w-full">
					{t("auth.signup.submit")}
				</SubmitButton>
				<ChallengeStatus checking={checking} />
			</form>
			<p className="mt-8 text-muted">
				{t("auth.signup.alreadyMember")}{" "}
				<Link href={l("/login")} className={authLink}>
					{t("auth.signup.logIn")}
				</Link>
			</p>
		</AuthFrame>
	);
}
