import type { ComponentProps } from "react";

import { cn } from "../../lib/utils";

/** The shared look of text controls: input, textarea and select all use it. */
export const inputClass =
	"flex h-10 w-full min-w-0 rounded-md border border-field bg-field-bg px-3 py-2 text-base text-fg shadow-xs transition-[color,box-shadow] outline-none placeholder:text-muted focus-visible:border-ring focus-visible:ring-[3px] focus-visible:ring-ring/50 disabled:cursor-not-allowed disabled:opacity-50 aria-invalid:border-danger aria-invalid:ring-danger/20 read-only:bg-surface read-only:text-muted md:text-sm";

export function Input({ className, type, ...props }: ComponentProps<"input">) {
	return <input type={type} className={cn(inputClass, className)} {...props} />;
}
