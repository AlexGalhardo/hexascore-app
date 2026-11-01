import { Head, useForm, usePage } from "@inertiajs/react";
import { type FormEvent, useState } from "react";

import Field from "../../components/Field";
import { useTurnstile } from "../../components/Turnstile";
import { useLocalized } from "../../lib/useLocalized";
import { useT } from "../../lib/useT";
import type { SharedProps } from "../../types";

/** Server maximum: ContactSchema in app/requests/contact.go (runes, after trimming). */
export const MESSAGE_MAX = 512;

interface Props {
	topics: string[];
	formToken: string;
}

// Contact is the public contact form. `website` is the honeypot: it is hidden from people and assistive tech.
export default function Index({ topics, formToken }: Props) {
	const t = useT();
	const l = useLocalized();
	const [sent, setSent] = useState(false);
	// Signed in: the server takes the sender from the session, so the fields only show who will be answered.
	const user = usePage<SharedProps>().props.auth?.user ?? null;
	const form = useForm({
		name: user ? (user.nickname ?? user.name) : "",
		email: user?.email ?? "",
		subject: topics[0] ?? "question",
		message: "",
		website: "",
		token: formToken,
	});

	const turnstile = useTurnstile("contact");

	const submit = (e: FormEvent) => {
		e.preventDefault();
		form.post(l("/contact"), {
			headers: turnstile.headers(),
			onSuccess: () => {
				setSent(true);
				form.reset("message");
			},
			onFinish: () => turnstile.reset(),
		});
	};

	return (
		<div className="mx-auto w-full max-w-xl">
			<Head title={t("contact.metaTitle")} />
			<h1 className="mb-2 text-2xl font-semibold">{t("contact.title")}</h1>
			<p className="mb-4 text-zinc-600 dark:text-zinc-400">{t("contact.body")}</p>
			{sent && <p role="status">{t("contact.sent")}</p>}
			<form onSubmit={submit} className="space-y-4">
				<Field
					label={t(user?.nickname ? "auth.fields.nickname" : "contact.name")}
					name="name"
					autoComplete={user ? "off" : "name"}
					readOnly={user !== null}
					value={form.data.name}
					onChange={(e) => form.setData("name", e.target.value)}
					error={form.errors.name}
				/>
				<Field
					label={t("contact.email")}
					name="email"
					type="email"
					autoComplete={user ? "off" : "email"}
					readOnly={user !== null}
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
						maxLength={MESSAGE_MAX}
						rows={6}
						name="message"
						aria-describedby="contact-message-hint contact-message-count"
						value={form.data.message}
						onChange={(e) => form.setData("message", e.target.value)}
						className="mt-1 block w-full rounded-md border border-zinc-300 bg-transparent px-3 py-2 dark:border-zinc-700"
					/>
					<span className="mt-1 flex justify-between gap-4 text-sm text-muted">
						<span id="contact-message-hint">{t("contact.hint")}</span>
						<span
							id="contact-message-count"
							data-testid="contact-message-count"
							className="font-mono tabular-nums"
						>
							{Array.from(form.data.message).length}/{MESSAGE_MAX}
						</span>
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
				{turnstile.render((form.errors as Record<string, string | undefined>).challenge)}
				<button
					type="submit"
					disabled={form.processing}
					className="rounded-md bg-primary px-4 py-2 font-medium text-on-primary hover:bg-primary-hover disabled:opacity-50"
				>
					{t("contact.submit")}
				</button>
			</form>
		</div>
	);
}
