"use client";

import { useTranslations } from "next-intl";
import { useState, useTransition } from "react";
import { addToList, createList, removeFromList } from "@/actions/lists";
import { buttonClass, Input, Notice } from "@/components/ui/form";
import type { ActionError } from "@/lib/action-result";
import { LIST_LIMITS, type ListType } from "@/lib/social/constants";

// `lists` are only the viewer's lists of the item's type, so their count is the quota used for that type.
type Props = { itemId: string; type: ListType; lists: { id: string; title: string; contains: boolean }[] };

export function AddToList({ itemId, type, lists }: Props) {
	const t = useTranslations("lists");
	const tAction = useTranslations("actions");
	const [entries, setEntries] = useState(lists);
	const [title, setTitle] = useState("");
	const [error, setError] = useState<ActionError | null>(null);
	const [pending, startTransition] = useTransition();
	const atLimit = entries.length >= LIST_LIMITS.perType;

	function toggle(listId: string, contains: boolean): void {
		setError(null);
		setEntries((prev) => prev.map((l) => (l.id === listId ? { ...l, contains: !contains } : l)));
		startTransition(async () => {
			const result = contains ? await removeFromList({ listId, itemId }) : await addToList({ listId, itemId });
			if (!result.ok) {
				setEntries((prev) => prev.map((l) => (l.id === listId ? { ...l, contains } : l)));
				setError(result.error);
			}
		});
	}

	function create(): void {
		setError(null);
		startTransition(async () => {
			const result = await createList({ title, description: "", isPublic: true, type }, itemId);
			if (!result.ok) return setError(result.error);
			setEntries((prev) => [{ id: result.id, title: title.trim(), contains: true }, ...prev]);
			setTitle("");
		});
	}

	return (
		<details className="group rounded-md border border-border">
			<summary className="flex min-h-11 cursor-pointer items-center px-4 text-sm font-medium">
				{t("addToList")}
			</summary>
			<div className="grid gap-3 border-t border-border p-3">
				{entries.length ? (
					<ul className="grid gap-1">
						{entries.map((l) => (
							<li key={l.id}>
								<label className="flex min-h-11 items-center gap-3 text-sm">
									<input
										type="checkbox"
										checked={l.contains}
										disabled={pending}
										onChange={() => toggle(l.id, l.contains)}
										className="size-5"
									/>
									<span className="truncate">{l.title}</span>
								</label>
							</li>
						))}
					</ul>
				) : (
					<p className="text-sm text-muted">{t("noListsOfType", { type })}</p>
				)}
				{atLimit ? (
					<Notice tone="info">{t("limitReached", { type, limit: LIST_LIMITS.perType })}</Notice>
				) : (
					<form
						className="flex gap-2"
						onSubmit={(e) => {
							e.preventDefault();
							create();
						}}
					>
						<Input
							aria-label={t("fieldTitle")}
							aria-describedby={`list-quota-${itemId}`}
							placeholder={t("newWithItem")}
							maxLength={LIST_LIMITS.title}
							required
							value={title}
							onChange={(e) => setTitle(e.target.value)}
						/>
						<button type="submit" disabled={pending || !title.trim()} className={buttonClass("secondary")}>
							{t("create")}
						</button>
					</form>
				)}
				<p id={`list-quota-${itemId}`} className="text-xs text-muted">
					{t("quota", { type, count: entries.length, limit: LIST_LIMITS.perType })}
				</p>
				{error ? (
					<Notice tone="error">
						{error === "limit" ? t("limitReached", { type, limit: LIST_LIMITS.perType }) : tAction(error)}
					</Notice>
				) : null}
			</div>
		</details>
	);
}
