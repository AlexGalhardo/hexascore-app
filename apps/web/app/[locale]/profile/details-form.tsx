"use client";

import { useTranslations } from "next-intl";
import { type FormEvent, useState, useTransition } from "react";
import { updateName } from "@/actions/account";
import { buttonClass, Field, Input, Notice } from "@/components/ui/form";
import type { ActionResult } from "@/lib/action-result";

type Props = { name: string; email: string; nickname: string | null };

export function DetailsForm({ name, email, nickname }: Props) {
	const t = useTranslations("account");
	const ta = useTranslations("actions");
	const [result, setResult] = useState<ActionResult | null>(null);
	const [pending, startTransition] = useTransition();

	function onSubmit(event: FormEvent<HTMLFormElement>): void {
		event.preventDefault();
		const value = new FormData(event.currentTarget).get("name");
		startTransition(async () => setResult(await updateName(value)));
	}

	return (
		<form onSubmit={onSubmit} className="flex flex-col gap-4" noValidate>
			{result &&
				(result.ok ? (
					<Notice tone="success">{t("saved")}</Notice>
				) : (
					<Notice tone="error">{ta(result.error)}</Notice>
				))}
			<dl className="grid grid-cols-[auto_1fr] gap-x-4 gap-y-1 text-sm">
				<dt className="text-muted">{t("details.email")}</dt>
				<dd className="break-all">{email}</dd>
				{nickname && (
					<>
						<dt className="text-muted">{t("details.nickname")}</dt>
						<dd className="font-mono">@{nickname}</dd>
					</>
				)}
			</dl>
			<Field id="profile-name" label={t("details.name")}>
				<Input
					id="profile-name"
					name="name"
					required
					maxLength={100}
					autoComplete="name"
					defaultValue={name}
					aria-invalid={result?.ok === false && result.error === "invalid"}
				/>
			</Field>
			<button type="submit" disabled={pending} className={buttonClass("secondary", "self-start")}>
				{t("details.save")}
			</button>
		</form>
	);
}
