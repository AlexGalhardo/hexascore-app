"use client";

import { useTranslations } from "next-intl";
import { useActionState, useState, useTransition } from "react";
import { sendContactMessage } from "@/actions/contact";
import { buttonClass, Field, Input, inputClass, Notice } from "@/components/ui/form";
import type { ActionResult } from "@/lib/action-result";
import { CONTACT_MESSAGE_MAX, CONTACT_MESSAGE_MIN, CONTACT_SUBJECTS } from "@/lib/validation";

type Props = { user: { name: string; email: string } | null };

export function ContactForm({ user }: Props) {
	const t = useTranslations("contact");
	const tAction = useTranslations("actions");
	const [state, action, pending] = useActionState<ActionResult | null, FormData>(sendContactMessage, null);
	const [length, setLength] = useState(0);
	const [, startTransition] = useTransition();

	if (state?.ok) return <Notice tone="success">{t("sent")}</Notice>;

	const invalid = (field: string): boolean => state?.ok === false && Boolean(state.fields?.includes(field));

	return (
		<form
			// Submitting through onSubmit (not `action`) keeps React from resetting the fields when the server rejects them.
			onSubmit={(event) => {
				event.preventDefault();
				const data = new FormData(event.currentTarget);
				startTransition(() => action(data));
			}}
			className="flex flex-col gap-4"
		>
			{state?.ok === false && <Notice tone="error">{tAction(state.error)}</Notice>}
			{user && <p className="text-sm text-muted">{t("loggedIn")}</p>}
			{/* Disabled fields are not submitted: the server takes name/email from the session for logged users. */}
			<Field id="contact-name" label={t("name")}>
				<Input
					id="contact-name"
					name="name"
					required
					maxLength={100}
					autoComplete="name"
					defaultValue={user?.name}
					disabled={Boolean(user)}
					aria-invalid={invalid("name")}
				/>
			</Field>
			<Field id="contact-email" label={t("email")}>
				<Input
					id="contact-email"
					name="email"
					type="email"
					required
					autoComplete="email"
					defaultValue={user?.email}
					disabled={Boolean(user)}
					aria-invalid={invalid("email")}
				/>
			</Field>
			<Field id="contact-subject" label={t("subject")}>
				<select
					id="contact-subject"
					name="subject"
					required
					defaultValue="question"
					className={inputClass}
					aria-invalid={invalid("subject")}
				>
					{CONTACT_SUBJECTS.map((s) => (
						<option key={s} value={s}>
							{t(`subjects.${s}`)}
						</option>
					))}
				</select>
			</Field>
			<Field id="contact-message" label={t("message")}>
				<textarea
					id="contact-message"
					name="message"
					rows={5}
					required
					minLength={CONTACT_MESSAGE_MIN}
					maxLength={CONTACT_MESSAGE_MAX}
					onChange={(e) => setLength(e.target.value.length)}
					aria-invalid={invalid("message")}
					aria-describedby="contact-message-count"
					className={`${inputClass} py-2`}
				/>
				<p id="contact-message-count" className="flex justify-between text-sm text-muted">
					<span>{t("minHint", { min: CONTACT_MESSAGE_MIN })}</span>
					<span className="font-mono tabular-nums">
						{length}/{CONTACT_MESSAGE_MAX}
					</span>
				</p>
			</Field>
			<button type="submit" disabled={pending} className={buttonClass("primary", "self-start")}>
				{t("submit")}
			</button>
		</form>
	);
}
