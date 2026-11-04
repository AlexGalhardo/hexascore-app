import { cva, type VariantProps } from "class-variance-authority";
import type { HTMLAttributes, ReactElement } from "react";

import { cn } from "../../../lib/cn";

const alertVariants = cva("flex items-start gap-2 rounded-lg border px-3 py-2.5 text-sm", {
	variants: {
		variant: {
			default: "border-border bg-surface text-fg",
			destructive: "border-danger/40 bg-danger/10 text-danger",
			success: "border-score-high/50 bg-score-high/10 text-fg",
		},
	},
	defaultVariants: { variant: "default" },
});

/** Errors use role="alert"; everything else is a polite status. */
export function Alert({
	className,
	variant,
	...props
}: HTMLAttributes<HTMLDivElement> & VariantProps<typeof alertVariants>): ReactElement {
	return (
		<div
			role={variant === "destructive" ? "alert" : "status"}
			className={cn(alertVariants({ variant }), className)}
			{...props}
		/>
	);
}
