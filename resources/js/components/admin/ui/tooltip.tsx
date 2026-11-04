import { Tooltip as BaseTooltip } from "@base-ui/react/tooltip";
import type { ReactElement, ReactNode } from "react";

import { cn } from "../../../lib/cn";

export const TooltipProvider = BaseTooltip.Provider;

/** A short label for an icon-only control; the control keeps its own aria-label. */
export function Tooltip({
	label,
	children,
	side = "right",
	className,
}: {
	label: string;
	children: ReactElement;
	side?: "top" | "right" | "bottom" | "left";
	className?: string;
}): ReactElement {
	return (
		<BaseTooltip.Root>
			<BaseTooltip.Trigger render={children} />
			<BaseTooltip.Portal>
				<BaseTooltip.Positioner side={side} sideOffset={6}>
					<BaseTooltip.Popup
						className={cn(
							"z-[70] rounded-md bg-fg px-2.5 py-1.5 text-xs font-medium text-bg shadow-md",
							className,
						)}
					>
						{label as ReactNode}
					</BaseTooltip.Popup>
				</BaseTooltip.Positioner>
			</BaseTooltip.Portal>
		</BaseTooltip.Root>
	);
}
