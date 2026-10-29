import { Head, useForm } from "@inertiajs/react";
import type { FormEvent } from "react";

import Field from "../../components/Field";
import { challengeHeaders } from "../../lib/challenge";
import { useT } from "../../lib/useT";

export default function Recover() {
	const t = useT();
	const form = useForm({ email: "" });

	const submit = async (e: FormEvent) => {
		e.preventDefault();
		form.post("/recover-password", { headers: await challengeHeaders("reset") });
	};

	return (
		<>
			<Head title={t("auth.recover.metaTitle")} />
			<h1 className="mb-2 text-2xl font-semibold">{t("auth.recover.title")}</h1>
			<p className="mb-4 text-zinc-600 dark:text-zinc-400">{t("auth.recover.body")}</p>
			<form onSubmit={submit} className="space-y-4">
				<Field
					label={t("auth.fields.email")}
					type="email"
					autoComplete="email"
					required
					value={form.data.email}
					onChange={(e) => form.setData("email", e.target.value)}
					error={form.errors.email}
				/>
				<button
					type="submit"
					disabled={form.processing}
					className="rounded-md bg-indigo-600 px-4 py-2 font-medium text-white hover:bg-indigo-500 disabled:opacity-50"
				>
					{t("auth.recover.submit")}
				</button>
			</form>
		</>
	);
}
