"use client";

import { useTranslations } from "next-intl";
import { useId, useState, useTransition } from "react";
import { saveItemStatus } from "@/actions/item-status";
import { inputClass, Notice } from "@/components/ui/form";
import type { ActionError } from "@/lib/action-result";
import type { ItemType } from "@/lib/db/schema/catalog";
import { ITEM_STATUSES, type ItemStatus } from "@/lib/social/constants";

type Props = { itemId: string; type: ItemType; initial: ItemStatus | null };

export function ItemStatusSelect({ itemId, type, initial }: Props) {
	const t = useTranslations("social.status");
	const tAction = useTranslations("actions");
	const id = useId();
	const [value, setValue] = useState<ItemStatus | null>(initial);
	const [error, setError] = useState<ActionError | null>(null);
	const [pending, startTransition] = useTransition();

	function onChange(next: ItemStatus | null): void {
		const previous = value;
		setValue(next);
		setError(null);
		startTransition(async () => {
			const result = await saveItemStatus({ itemId, status: next });
			if (!result.ok) {
				setValue(previous);
				setError(result.error);
			}
		});
	}

	return (
		<div className="grid gap-1">
			<label htmlFor={id} className="text-sm font-medium">
				{t("label")}
			</label>
			<select
				id={id}
				value={value ?? ""}
				disabled={pending}
				onChange={(e) => onChange((e.target.value || null) as ItemStatus | null)}
				className={`${inputClass} sm:w-60`}
			>
				<option value="">{t("none")}</option>
				{ITEM_STATUSES.map((s) => (
					<option key={s} value={s}>
						{t(`${type}.${s}`)}
					</option>
				))}
			</select>
			{error ? <Notice tone="error">{tAction(error)}</Notice> : null}
		</div>
	);
}
