import { Slot } from "@radix-ui/react-slot";
import { cva, type VariantProps } from "class-variance-authority";
import type { ComponentProps } from "react";

import { cn } from "../../lib/utils";

// shadcn/ui button (new-york), tuned to the HexaScore tokens: flat, 1 px hairlines, 6 px radius, 2 px focus ring.
export const buttonVariants = cva(
	"inline-flex shrink-0 items-center justify-center gap-2 rounded-md text-sm font-medium whitespace-nowrap transition-colors outline-none focus-visible:ring-[3px] focus-visible:ring-ring/50 disabled:pointer-events-none disabled:opacity-50 aria-disabled:pointer-events-none aria-disabled:opacity-50 [&_svg]:pointer-events-none [&_svg]:shrink-0 [&_svg:not([class*='size-'])]:size-4",
	{
		variants: {
			variant: {
				default: "bg-primary text-on-primary shadow-xs hover:bg-primary-hover",
				destructive: "bg-danger text-on-primary shadow-xs hover:bg-danger/90",
				outline: "border border-border bg-bg shadow-xs hover:bg-hover",
				secondary: "bg-surface text-fg hover:bg-hover",
				ghost: "hover:bg-hover",
				link: "text-fg underline-offset-4 hover:underline",
			},
			size: {
				default: "h-10 px-4 py-2",
				sm: "h-9 gap-1.5 px-3",
				lg: "h-11 px-6 text-[15px]",
				icon: "size-10",
			},
		},
		defaultVariants: { variant: "default", size: "default" },
	},
);

export function Button({
	className,
	variant,
	size,
	asChild = false,
	...props
}: ComponentProps<"button"> & VariantProps<typeof buttonVariants> & { asChild?: boolean }) {
	const Comp = asChild ? Slot : "button";
	return <Comp className={cn(buttonVariants({ variant, size }), className)} {...props} />;
}
