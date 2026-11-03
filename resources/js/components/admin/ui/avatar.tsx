import { Avatar as BaseAvatar } from "@base-ui/react/avatar";
import type { ReactElement } from "react";

import { cn } from "../../../lib/cn";

/** Initials in a circle; the picture is optional and falls back to the initials. */
export function Avatar({
	name,
	src,
	className,
}: {
	name: string;
	src?: string;
	className?: string;
}): ReactElement {
	const initials = name
		.split(/\s+/)
		.filter(Boolean)
		.slice(0, 2)
		.map((part) => part[0]?.toUpperCase())
		.join("");
	return (
		<BaseAvatar.Root
			className={cn(
				"inline-flex size-8 shrink-0 items-center justify-center overflow-hidden rounded-full bg-surface text-xs font-semibold text-fg",
				className,
			)}
		>
			{src && <BaseAvatar.Image src={src} alt="" className="size-full object-cover" />}
			<BaseAvatar.Fallback>{initials || "?"}</BaseAvatar.Fallback>
		</BaseAvatar.Root>
	);
}
