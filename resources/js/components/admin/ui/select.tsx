import type { ReactElement, SelectHTMLAttributes } from "react";

import { cn } from "../../../lib/cn";

/** A native select styled like Input: the platform control keeps keyboard, mobile pickers and form posts. */
export function Select({
	className,
	...props
}: SelectHTMLAttributes<HTMLSelectElement>): ReactElement {
	return (
		<select
			className={cn(
				"h-9 w-full min-w-0 rounded-md border border-field bg-field-bg px-3 text-sm text-fg outline-none transition-colors focus-visible:ring-2 focus-visible:ring-primary disabled:cursor-not-allowed disabled:opacity-50",
				className,
			)}
			{...props}
		/>
	);
}
