"use client";

import { useTranslations } from "next-intl";
import { useState, useTransition } from "react";
import { setProfilePrivate } from "@/actions/account";
import { Notice } from "@/components/ui/form";
import type { ActionError } from "@/lib/action-result";

export function PrivacyToggle({ initial }: { initial: boolean }) {
	const t = useTranslations("account.privacy");
	const tAction = useTranslations("actions");
	const [checked, setChecked] = useState(initial);
	const [error, setError] = useState<ActionError | null>(null);
	const [pending, startTransition] = useTransition();

	function onChange(value: boolean): void {
		setChecked(value);
		setError(null);
		startTransition(async () => {
			const result = await setProfilePrivate(value);
			if (!result.ok) {
				setChecked(!value);
				setError(result.error);
			}
		});
	}

	return (
		<div className="flex flex-col gap-2">
			{error && <Notice tone="error">{tAction(error)}</Notice>}
			<label className="flex min-h-11 items-center gap-3">
				<input
					type="checkbox"
					role="switch"
					checked={checked}
					aria-checked={checked}
					disabled={pending}
					onChange={(e) => onChange(e.target.checked)}
					aria-describedby="privacy-help"
					className="size-5"
				/>
				<span className="font-medium">{t("label")}</span>
			</label>
			<p id="privacy-help" className="text-sm text-muted">
				{t("help")}
			</p>
		</div>
	);
}
