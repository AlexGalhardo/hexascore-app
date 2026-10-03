"use client";

import { useTranslations } from "next-intl";
import { useState, useTransition } from "react";
import { openBillingPortal } from "@/actions/billing";

export function ManageBillingButton() {
	const t = useTranslations("billing.checkout");
	const [pending, start] = useTransition();
	const [error, setError] = useState<string | null>(null);

	function open(): void {
		setError(null);
		start(async () => {
			const result = await openBillingPortal();
			if (result.ok) window.location.assign(result.url);
			else setError(t(`errors.${result.error}`));
		});
	}

	return (
		<div className="space-y-2">
			<button
				type="button"
				onClick={open}
				disabled={pending}
				className="min-h-11 rounded-md border border-border px-4 text-sm disabled:opacity-50"
			>
				{pending ? t("working") : t("manage")}
			</button>
			{error && (
				<p role="alert" className="text-sm text-score-low">
					{error}
				</p>
			)}
		</div>
	);
}
