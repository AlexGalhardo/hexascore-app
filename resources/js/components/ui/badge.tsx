import { Slot } from "@radix-ui/react-slot";
import { cva, type VariantProps } from "class-variance-authority";
import type { ComponentProps } from "react";

import { cn } from "../../lib/utils";

export const badgeVariants = cva(
	"inline-flex w-fit shrink-0 items-center justify-center gap-1 rounded-full border px-2.5 py-0.5 text-xs font-medium whitespace-nowrap [&>svg]:size-3 [&>svg]:pointer-events-none",
	{
		variants: {
			variant: {
				default: "border-transparent bg-primary text-on-primary",
				secondary: "border-transparent bg-surface text-fg",
				outline: "border-border text-fg",
				destructive: "border-transparent bg-danger text-on-primary",
			},
		},
		defaultVariants: { variant: "default" },
	},
);

export function Badge({
	className,
	variant,
	asChild = false,
	...props
}: ComponentProps<"span"> & VariantProps<typeof badgeVariants> & { asChild?: boolean }) {
	const Comp = asChild ? Slot : "span";
	return <Comp className={cn(badgeVariants({ variant }), className)} {...props} />;
}
