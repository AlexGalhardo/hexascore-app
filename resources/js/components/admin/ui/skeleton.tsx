import type { HTMLAttributes, ReactElement } from "react";

import { cn } from "../../../lib/cn";

export function Skeleton({ className, ...props }: HTMLAttributes<HTMLDivElement>): ReactElement {
	return (
		<div
			aria-hidden="true"
			className={cn(
				"animate-pulse rounded-md bg-surface motion-reduce:animate-none",
				className,
			)}
			{...props}
		/>
	);
}
