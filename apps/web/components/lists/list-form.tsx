"use client";

import { useTranslations } from "next-intl";
import { useId, useState, useTransition } from "react";
import { createList, updateList } from "@/actions/lists";
import { buttonClass, Field, Input, inputClass, Notice } from "@/components/ui/form";
import type { ActionError } from "@/lib/action-result";
import { LIST_LIMITS } from "@/lib/social/constants";

type Values = { title: string; description: string; isPublic: boolean };
type Props = { listId?: string; initial?: Values };

/** Creates a list (no `listId`) or edits one. */
export function ListForm({ listId, initial }: Props) {
	const t = useTranslations("lists");
	const tAction = useTranslations("actions");
	const id = useId();
	const [values, setValues] = useState<Values>(initial ?? { title: "", description: "", isPublic: true });
	const [result, setResult] = useState<{ error: ActionError; fields?: string[] } | "saved" | null>(null);
	const [pending, startTransition] = useTransition();

	function submit(): void {
		setResult(null);
		startTransition(async () => {
			const response = listId ? await updateList(listId, values) : await createList(values);
			if (!response.ok) return setResult({ error: response.error, fields: response.fields });
			setResult("saved");
			if (!listId) setValues({ title: "", description: "", isPublic: true });
		});
	}

	const fieldError = (name: string) =>
		result && result !== "saved" && result.fields?.includes(name) ? t(`errors.${name}`) : undefined;

	return (
		<form
			className="grid max-w-xl gap-3"
			onSubmit={(e) => {
				e.preventDefault();
				submit();
			}}
		>
			<Field id={`${id}-title`} label={t("fieldTitle")} error={fieldError("title")}>
				<Input
					id={`${id}-title`}
					required
					maxLength={LIST_LIMITS.title}
					value={values.title}
					aria-invalid={Boolean(fieldError("title"))}
					onChange={(e) => setValues({ ...values, title: e.target.value })}
				/>
			</Field>
			<Field
				id={`${id}-description`}
				label={t("fieldDescription")}
				hint={`${values.description.length}/${LIST_LIMITS.description}`}
				error={fieldError("description")}
			>
				<textarea
					id={`${id}-description`}
					rows={3}
					maxLength={LIST_LIMITS.description}
					value={values.description}
					onChange={(e) => setValues({ ...values, description: e.target.value })}
					className={`${inputClass} py-2`}
				/>
			</Field>
			<label className="flex min-h-11 items-center gap-3 text-sm">
				<input
					type="checkbox"
					checked={values.isPublic}
					onChange={(e) => setValues({ ...values, isPublic: e.target.checked })}
					className="size-5"
				/>
				{t("publicLabel")}
			</label>
			{result === "saved" ? <Notice tone="success">{listId ? t("saved") : t("created")}</Notice> : null}
			{result && result !== "saved" ? <Notice tone="error">{tAction(result.error)}</Notice> : null}
			<button type="submit" disabled={pending} className={buttonClass("primary", "w-fit")}>
				{listId ? t("save") : t("create")}
			</button>
		</form>
	);
}
