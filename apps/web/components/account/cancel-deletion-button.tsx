"use client";

import { useTranslations } from "next-intl";
import { useState, useTransition } from "react";
import { cancelDeletion } from "@/actions/account";
import { buttonClass } from "@/components/ui/form";
import type { ActionError } from "@/lib/action-result";

export function CancelDeletionButton({ label }: { label: string }) {
	const tAction = useTranslations("actions");
	const [error, setError] = useState<ActionError | null>(null);
	const [pending, startTransition] = useTransition();

	return (
		<>
			{error && (
				<span role="alert" className="text-danger">
					{tAction(error)}
				</span>
			)}
			<button
				type="button"
				disabled={pending}
				onClick={() =>
					startTransition(async () => {
						const result = await cancelDeletion();
						if (!result.ok) setError(result.error);
					})
				}
				className={buttonClass("primary")}
			>
				{label}
			</button>
		</>
	);
}
