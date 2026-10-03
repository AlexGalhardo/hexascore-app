"use client";

import { type ReactNode, useActionState } from "react";
import type { ActionResult } from "@/lib/action-result";

type Props = {
	action: (prev: ActionResult | null, form: FormData) => Promise<ActionResult>;
	children: ReactNode;
	className?: string;
	confirm?: string;
};

/** Plain form for /admin Server Actions; shows the result code inline (admin is EN-only, no translation). */
export function ActionForm({ action, children, className, confirm }: Props) {
	const [state, formAction, pending] = useActionState(action, null);
	return (
		<form
			action={formAction}
			className={className ?? "flex flex-wrap items-center gap-2"}
			onSubmit={(e) => {
				if (confirm && !window.confirm(confirm)) e.preventDefault();
			}}
		>
			<fieldset disabled={pending} className="contents">
				{children}
			</fieldset>
			{state && !state.ok ? (
				<span role="alert" className="text-xs text-danger">
					{state.error}
					{state.fields?.length ? ` (${state.fields.join(", ")})` : ""}
				</span>
			) : null}
			{state?.ok ? (
				<span role="status" className="text-xs text-muted">
					done
				</span>
			) : null}
		</form>
	);
}
