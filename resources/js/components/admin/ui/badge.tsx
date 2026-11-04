import { cva, type VariantProps } from "class-variance-authority";
import type { HTMLAttributes, ReactElement } from "react";

import { cn } from "../../../lib/cn";

// The status colours reuse the score bands (tokens), so light and dark stay in one place.
export const badgeVariants = cva(
	"inline-flex w-fit shrink-0 items-center gap-1 rounded-md border px-2 py-0.5 text-xs font-medium whitespace-nowrap [&_svg]:size-3",
	{
		variants: {
			variant: {
				default: "border-transparent bg-fg text-bg",
				secondary: "border-transparent bg-surface text-fg",
				outline: "border-border text-fg",
				success: "border-transparent bg-score-high text-on-score",
				warning: "border-transparent bg-score-mid text-on-score",
				destructive: "border-transparent bg-score-low text-on-score",
			},
		},
		defaultVariants: { variant: "secondary" },
	},
);

export type BadgeProps = HTMLAttributes<HTMLSpanElement> & VariantProps<typeof badgeVariants>;

export function Badge({ className, variant, ...props }: BadgeProps): ReactElement {
	return <span className={cn(badgeVariants({ variant }), className)} {...props} />;
}
