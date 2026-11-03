import type { InputHTMLAttributes, ReactElement } from "react";

import { cn } from "../../../lib/cn";

export function Input({
	className,
	type = "text",
	...props
}: InputHTMLAttributes<HTMLInputElement>): ReactElement {
	return (
		<input
			type={type}
			className={cn(
				"h-9 w-full min-w-0 rounded-md border border-field bg-field-bg px-3 text-sm text-fg outline-none transition-colors placeholder:text-muted focus-visible:ring-2 focus-visible:ring-primary disabled:cursor-not-allowed disabled:opacity-50",
				className,
			)}
			{...props}
		/>
	);
}
