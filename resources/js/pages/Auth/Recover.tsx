import { Head, useForm } from "@inertiajs/react";
import { type FormEvent, useState } from "react";

import AuthFrame from "../../components/auth/AuthFrame";
import ChallengeStatus from "../../components/ChallengeStatus";
import Field from "../../components/Field";
import SubmitButton from "../../components/SubmitButton";
import { challengeHeaders } from "../../lib/challenge";
import { useLocalized } from "../../lib/useLocalized";
import { useT } from "../../lib/useT";

export default function Recover() {
	const t = useT();
	const l = useLocalized();
	const form = useForm({ email: "" });
	const [checking, setChecking] = useState(false);

	const submit = async (e: FormEvent) => {
		e.preventDefault();
		form.post(l("/recover-password"), {
			headers: await challengeHeaders("reset", setChecking),
		});
	};

	return (
		<AuthFrame
			heading={t("auth.recover.heading")}
			lede={t("auth.recover.lede")}
			title={t("auth.recover.title")}
		>
			<Head title={t("auth.recover.metaTitle")} />
			<p className="mb-5 text-muted">{t("auth.recover.body")}</p>
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
				<SubmitButton pending={form.processing || checking} className="w-full">
					{t("auth.recover.submit")}
				</SubmitButton>
				<ChallengeStatus checking={checking} />
			</form>
		</AuthFrame>
	);
}
