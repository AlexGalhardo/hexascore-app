import type { ReactElement, TextareaHTMLAttributes } from "react";

import { cn } from "../../../lib/cn";

export function Textarea({
	className,
	...props
}: TextareaHTMLAttributes<HTMLTextAreaElement>): ReactElement {
	return (
		<textarea
			className={cn(
				"min-h-20 w-full rounded-md border border-field bg-field-bg px-3 py-2 text-sm text-fg outline-none transition-colors placeholder:text-muted focus-visible:ring-2 focus-visible:ring-primary disabled:cursor-not-allowed disabled:opacity-50",
				className,
			)}
			{...props}
		/>
	);
}
