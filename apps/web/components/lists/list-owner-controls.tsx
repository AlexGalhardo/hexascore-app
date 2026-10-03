"use client";

import { useLocale, useTranslations } from "next-intl";
import { useState, useTransition } from "react";
import { deleteList, moveInList, removeFromList } from "@/actions/lists";
import { buttonClass, Notice } from "@/components/ui/form";
import { getPathname } from "@/i18n/navigation";
import type { ActionError } from "@/lib/action-result";

const small = "min-h-11 min-w-11 rounded-md border border-border px-2 text-sm hover:bg-surface disabled:opacity-50";

/** Move up/down and remove buttons for one entry of the owner's list. */
export function ListEntryControls({
	listId,
	itemId,
	title,
	first,
	last,
}: {
	listId: string;
	itemId: string;
	title: string;
	first: boolean;
	last: boolean;
}) {
	const t = useTranslations("lists");
	const tAction = useTranslations("actions");
	const [error, setError] = useState<ActionError | null>(null);
	const [pending, startTransition] = useTransition();

	function run(action: () => Promise<{ ok: boolean; error?: ActionError }>): void {
		setError(null);
		startTransition(async () => {
			const result = await action();
			if (!result.ok && result.error) setError(result.error);
		});
	}

	return (
		<div className="flex flex-wrap items-center gap-1">
			<button
				type="button"
				disabled={pending || first}
				aria-label={t("moveUpLabel", { title })}
				onClick={() => run(() => moveInList({ listId, itemId, direction: "up" }))}
				className={small}
			>
				↑
			</button>
			<button
				type="button"
				disabled={pending || last}
				aria-label={t("moveDownLabel", { title })}
				onClick={() => run(() => moveInList({ listId, itemId, direction: "down" }))}
				className={small}
			>
				↓
			</button>
			<button
				type="button"
				disabled={pending}
				aria-label={t("removeLabel", { title })}
				onClick={() => run(() => removeFromList({ listId, itemId }))}
				className={small}
			>
				{t("remove")}
			</button>
			{error ? <Notice tone="error">{tAction(error)}</Notice> : null}
		</div>
	);
}

export function DeleteListButton({ listId, nickname }: { listId: string; nickname: string }) {
	const t = useTranslations("lists");
	const tAction = useTranslations("actions");
	const locale = useLocale();
	const [error, setError] = useState<ActionError | null>(null);
	const [pending, startTransition] = useTransition();

	function onClick(): void {
		if (!window.confirm(t("deleteConfirm"))) return;
		startTransition(async () => {
			const result = await deleteList(listId);
			if (!result.ok) return setError(result.error);
			window.location.assign(
				getPathname({ href: { pathname: `/u/${nickname}`, query: { tab: "lists" } }, locale }),
			);
		});
	}

	return (
		<div className="grid gap-2">
			<button type="button" disabled={pending} onClick={onClick} className={buttonClass("danger", "w-fit")}>
				{t("delete")}
			</button>
			{error ? <Notice tone="error">{tAction(error)}</Notice> : null}
		</div>
	);
}
