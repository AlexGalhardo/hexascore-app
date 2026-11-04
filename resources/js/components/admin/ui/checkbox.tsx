import type { InputHTMLAttributes, ReactElement } from "react";

import { cn } from "../../../lib/cn";

/** A native checkbox: it posts with the form and needs no state of its own. */
export function Checkbox({
	className,
	...props
}: Omit<InputHTMLAttributes<HTMLInputElement>, "type">): ReactElement {
	return (
		<input
			type="checkbox"
			className={cn(
				"size-4 shrink-0 rounded border-field accent-primary outline-none focus-visible:ring-2 focus-visible:ring-primary disabled:opacity-50",
				className,
			)}
			{...props}
		/>
	);
}
