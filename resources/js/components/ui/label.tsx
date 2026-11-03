import type { ComponentProps } from "react";

import { cn } from "../../lib/utils";

export function Label({ className, ...props }: ComponentProps<"label">) {
	return (
		// biome-ignore lint/a11y/noLabelWithoutControl: the caller passes htmlFor or wraps the control
		<label
			className={cn("text-sm leading-none font-medium select-none", className)}
			{...props}
		/>
	);
}
