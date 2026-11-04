import type { HTMLAttributes, ReactElement } from "react";

import { cn } from "../../../lib/cn";

type Props = HTMLAttributes<HTMLDivElement>;

export function Card({ className, ...props }: Props): ReactElement {
	return (
		<div
			className={cn(
				"flex flex-col gap-4 rounded-xl border border-border bg-bg py-5 text-fg",
				className,
			)}
			{...props}
		/>
	);
}

export function CardHeader({ className, ...props }: Props): ReactElement {
	return <div className={cn("grid gap-1.5 px-5", className)} {...props} />;
}

export function CardTitle({ className, ...props }: Props): ReactElement {
	return <div className={cn("text-sm leading-none font-semibold", className)} {...props} />;
}

export function CardDescription({ className, ...props }: Props): ReactElement {
	return <div className={cn("text-sm text-muted", className)} {...props} />;
}

export function CardContent({ className, ...props }: Props): ReactElement {
	return <div className={cn("px-5", className)} {...props} />;
}

export function CardFooter({ className, ...props }: Props): ReactElement {
	return <div className={cn("flex items-center px-5", className)} {...props} />;
}
