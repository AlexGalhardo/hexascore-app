import { Dialog as BaseDialog } from "@base-ui/react/dialog";
import type { ReactNode } from "react";

export type DialogSize = "sm" | "md" | "lg" | "xl";

const WIDTHS: Record<DialogSize, string> = {
	sm: "sm:max-w-sm",
	md: "sm:max-w-md",
	lg: "sm:max-w-2xl",
	xl: "sm:max-w-5xl",
};

// Every dialog is centred on both axes, full width minus margins on mobile and
// above the navbar (z-40) and the cookie banner (z-50) (SDD 09 §7.5). Dialogs
// take their popup class from here, never their own.
export function dialogPopup(size: DialogSize = "md"): string {
	return `fixed top-1/2 left-1/2 z-[60] w-[calc(100%-2rem)] ${WIDTHS[size]} -translate-x-1/2 -translate-y-1/2 rounded-lg border border-border bg-bg p-6 text-fg shadow-lg outline-none`;
}

export const dialogBackdrop = "fixed inset-0 z-[55] bg-black/50";

export default function Dialog({
	open,
	onOpenChange,
	title,
	description,
	size = "md",
	children,
}: {
	open: boolean;
	onOpenChange: (open: boolean) => void;
	title: string;
	description?: string;
	size?: DialogSize;
	children: ReactNode;
}) {
	return (
		<BaseDialog.Root open={open} onOpenChange={onOpenChange}>
			<BaseDialog.Portal>
				<BaseDialog.Backdrop className={dialogBackdrop} />
				<BaseDialog.Popup className={dialogPopup(size)}>
					<BaseDialog.Title className="text-lg font-semibold">{title}</BaseDialog.Title>
					{description && (
						<BaseDialog.Description className="mt-1 text-sm text-muted">
							{description}
						</BaseDialog.Description>
					)}
					{children}
				</BaseDialog.Popup>
			</BaseDialog.Portal>
		</BaseDialog.Root>
	);
}
