"use client";

import { useTranslations } from "next-intl";
import { useId, useState, useTransition } from "react";
import { createList, updateList } from "@/actions/lists";
import { buttonClass, Field, Input, inputClass, Notice } from "@/components/ui/form";
import type { ActionError } from "@/lib/action-result";
import { LIST_LIMITS, LIST_TYPES, type ListType } from "@/lib/social/constants";

type Values = { title: string; description: string; isPublic: boolean };
/** Edit mode: `listId` + `initial`. Create mode: `counts` = the owner's lists per type (quota display). */
type Props =
	| { listId: string; initial: Values; counts?: never }
	| { listId?: never; initial?: never; counts: Record<ListType, number> };

const empty: Values = { title: "", description: "", isPublic: true };

/** Creates a list (no `listId`, the category is required) or edits one (the category cannot change). */
export function ListForm({ listId, initial, counts }: Props) {
	const t = useTranslations("lists");
	const tAction = useTranslations("actions");
	const tType = useTranslations("itemTypes");
	const id = useId();
	const [values, setValues] = useState<Values>(initial ?? empty);
	const [used, setUsed] = useState(counts);
	const [type, setType] = useState<ListType>(
		() => LIST_TYPES.find((ty) => (counts?.[ty] ?? 0) < LIST_LIMITS.perType) ?? LIST_TYPES[0],
	);
	const [result, setResult] = useState<{ error: ActionError; fields?: string[] } | "saved" | null>(null);
	const [pending, startTransition] = useTransition();
	const atLimit = used ? used[type] >= LIST_LIMITS.perType : false;

	function submit(): void {
		setResult(null);
		startTransition(async () => {
			const response = listId ? await updateList(listId, values) : await createList({ ...values, type });
			if (!response.ok) return setResult({ error: response.error, fields: response.fields });
			setResult("saved");
			if (!listId) {
				setValues(empty);
				setUsed((prev) => (prev ? { ...prev, [type]: prev[type] + 1 } : prev));
			}
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
			{used ? (
				<Field
					id={`${id}-type`}
					label={t("fieldType")}
					hint={t("quota", { type, count: used[type], limit: LIST_LIMITS.perType })}
					error={fieldError("type")}
				>
					<select
						id={`${id}-type`}
						required
						value={type}
						onChange={(e) => {
							setResult(null);
							setType(e.target.value as ListType);
						}}
						className={`${inputClass} sm:w-60`}
					>
						{LIST_TYPES.map((ty) => (
							<option key={ty} value={ty}>
								{tType(ty)}
							</option>
						))}
					</select>
				</Field>
			) : null}
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
			{atLimit ? <Notice tone="info">{t("limitReached", { type, limit: LIST_LIMITS.perType })}</Notice> : null}
			{result === "saved" ? <Notice tone="success">{listId ? t("saved") : t("created")}</Notice> : null}
			{result && result !== "saved" ? (
				<Notice tone="error">
					{result.error === "limit" && !listId
						? t("limitReached", { type, limit: LIST_LIMITS.perType })
						: tAction(result.error)}
				</Notice>
			) : null}
			<button type="submit" disabled={pending || atLimit} className={buttonClass("primary", "w-fit")}>
				{listId ? t("save") : t("create")}
			</button>
		</form>
	);
}
