import { Head, useForm } from "@inertiajs/react";
import { type FormEvent, useState } from "react";

import Field from "../../components/Field";
import { useT } from "../../lib/useT";

interface Props {
	topics: string[];
	formToken: string;
}

// Contact is the public contact form. `website` is the honeypot: it is hidden from people and assistive tech.
export default function Index({ topics, formToken }: Props) {
	const t = useT();
	const [sent, setSent] = useState(false);
	const form = useForm({
		name: "",
		email: "",
		subject: topics[0] ?? "question",
		message: "",
		website: "",
		token: formToken,
	});

	const submit = (e: FormEvent) => {
		e.preventDefault();
		form.post("/contact", {
			onSuccess: () => {
				setSent(true);
				form.reset("message");
			},
		});
	};

	return (
		<>
			<Head title={t("contact.metaTitle")} />
			<h1 className="mb-2 text-2xl font-semibold">{t("contact.title")}</h1>
			<p className="mb-4 text-zinc-600 dark:text-zinc-400">{t("contact.body")}</p>
			{sent && <p role="status">{t("contact.sent")}</p>}
			<form onSubmit={submit} className="space-y-4">
				<Field
					label={t("contact.name")}
					autoComplete="name"
					value={form.data.name}
					onChange={(e) => form.setData("name", e.target.value)}
					error={form.errors.name}
				/>
				<Field
					label={t("contact.email")}
					type="email"
					autoComplete="email"
					value={form.data.email}
					onChange={(e) => form.setData("email", e.target.value)}
					error={form.errors.email}
				/>
				<label className="block">
					<span className="font-medium">{t("contact.subject")}</span>
					<select
						value={form.data.subject}
						onChange={(e) => form.setData("subject", e.target.value)}
						className="mt-1 block w-full rounded-md border border-zinc-300 bg-transparent px-3 py-2 dark:border-zinc-700"
					>
						{topics.map((topic) => (
							<option key={topic} value={topic}>
								{t(`contact.topics.${topic}`)}
							</option>
						))}
					</select>
				</label>
				<label className="block">
					<span className="font-medium">{t("contact.message")}</span>
					<textarea
						required
						minLength={32}
						maxLength={512}
						rows={6}
						value={form.data.message}
						onChange={(e) => form.setData("message", e.target.value)}
						className="mt-1 block w-full rounded-md border border-zinc-300 bg-transparent px-3 py-2 dark:border-zinc-700"
					/>
					<span className="text-sm text-zinc-600 dark:text-zinc-400">
						{t("contact.hint")}
					</span>
					{form.errors.message && (
						<p className="mt-1 text-sm text-red-600">{form.errors.message}</p>
					)}
				</label>
				<div aria-hidden="true" className="absolute -left-[9999px]">
					<input
						tabIndex={-1}
						autoComplete="off"
						name="website"
						value={form.data.website}
						onChange={(e) => form.setData("website", e.target.value)}
					/>
				</div>
				<button
					type="submit"
					disabled={form.processing}
					className="rounded-md bg-indigo-600 px-4 py-2 font-medium text-white hover:bg-indigo-500 disabled:opacity-50"
				>
					{t("contact.submit")}
				</button>
			</form>
		</>
	);
}
