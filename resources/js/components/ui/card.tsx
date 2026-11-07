import type { ComponentProps } from "react";

import { cn } from "../../lib/utils";

export function Card({ className, ...props }: ComponentProps<"div">) {
	return (
		<div
			className={cn(
				"flex flex-col gap-4 rounded-lg border border-border bg-surface p-5 text-fg shadow-xs",
				className,
			)}
			{...props}
		/>
	);
}

export function CardHeader({ className, ...props }: ComponentProps<"div">) {
	return <div className={cn("grid gap-1.5", className)} {...props} />;
}

export function CardTitle({ className, ...props }: ComponentProps<"h3">) {
	// biome-ignore lint/a11y/useHeadingContent: the caller passes the children
	return (
		<h3
			className={cn("text-lg leading-none font-semibold tracking-tight", className)}
			{...props}
		/>
	);
}

export function CardDescription({ className, ...props }: ComponentProps<"p">) {
	return <p className={cn("text-sm text-muted", className)} {...props} />;
}

export function CardContent({ className, ...props }: ComponentProps<"div">) {
	return <div className={cn(className)} {...props} />;
}

export function CardFooter({ className, ...props }: ComponentProps<"div">) {
	return <div className={cn("flex items-center gap-2", className)} {...props} />;
}
