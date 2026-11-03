import { Dialog } from "@base-ui/react/dialog";
import type { ReactElement, ReactNode } from "react";

import { cn } from "../../../lib/cn";

/** A side panel (the mobile sidebar). Built on Base UI Dialog: focus trap, Escape and inert page. */
export function Sheet({
	open,
	onOpenChange,
	title,
	side = "left",
	className,
	children,
}: {
	open: boolean;
	onOpenChange: (open: boolean) => void;
	title: string;
	side?: "left" | "right";
	className?: string;
	children: ReactNode;
}): ReactElement {
	return (
		<Dialog.Root open={open} onOpenChange={onOpenChange}>
			<Dialog.Portal>
				<Dialog.Backdrop className="fixed inset-0 z-[55] bg-black/50 transition-opacity data-[ending-style]:opacity-0 data-[starting-style]:opacity-0 motion-reduce:transition-none" />
				<Dialog.Popup
					className={cn(
						"fixed inset-y-0 z-[60] flex w-72 max-w-[85vw] flex-col border-border bg-bg text-fg shadow-lg outline-none transition-transform motion-reduce:transition-none",
						side === "left"
							? "left-0 border-r data-[ending-style]:-translate-x-full data-[starting-style]:-translate-x-full"
							: "right-0 border-l data-[ending-style]:translate-x-full data-[starting-style]:translate-x-full",
						className,
					)}
				>
					<Dialog.Title className="sr-only">{title}</Dialog.Title>
					{children}
				</Dialog.Popup>
			</Dialog.Portal>
		</Dialog.Root>
	);
}
