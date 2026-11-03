import { cva, type VariantProps } from "class-variance-authority";
import type { ButtonHTMLAttributes, ReactElement } from "react";

import { cn } from "../../../lib/cn";

export const buttonVariants = cva(
	"inline-flex shrink-0 items-center justify-center gap-2 rounded-md text-sm font-medium whitespace-nowrap transition-colors outline-none focus-visible:ring-2 focus-visible:ring-primary focus-visible:ring-offset-2 focus-visible:ring-offset-bg disabled:pointer-events-none disabled:opacity-50 [&_svg]:pointer-events-none [&_svg]:size-4 [&_svg]:shrink-0",
	{
		variants: {
			variant: {
				default: "bg-fg text-bg hover:bg-fg/90",
				primary: "bg-primary text-on-primary hover:bg-primary-hover",
				secondary: "bg-surface text-fg hover:bg-hover",
				outline: "border border-border bg-bg text-fg hover:bg-hover",
				ghost: "text-fg hover:bg-hover",
				destructive: "bg-danger text-white hover:bg-danger/90",
				link: "text-primary underline-offset-4 hover:underline",
			},
			size: {
				default: "h-9 px-4 py-2",
				sm: "h-8 gap-1.5 px-3 text-[13px]",
				lg: "h-10 px-6",
				icon: "size-9",
			},
		},
		defaultVariants: { variant: "default", size: "default" },
	},
);

export type ButtonProps = ButtonHTMLAttributes<HTMLButtonElement> &
	VariantProps<typeof buttonVariants>;

export function Button({
	className,
	variant,
	size,
	type = "button",
	...props
}: ButtonProps): ReactElement {
	return (
		<button
			type={type}
			className={cn(buttonVariants({ variant, size }), className)}
			{...props}
		/>
	);
}
