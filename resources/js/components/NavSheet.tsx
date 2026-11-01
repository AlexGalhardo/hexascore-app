import { Dialog as BaseDialog } from "@base-ui/react/dialog";
import type { ReactElement, ReactNode, RefObject } from "react";

import { Icon, iconButton } from "./Nav";

interface Props {
	open: boolean;
	onOpenChange: (open: boolean) => void;
	sheetRef: RefObject<HTMLDivElement | null>;
	finalFocus: RefObject<HTMLButtonElement | null>;
	title: string;
	closeLabel: string;
	children: ReactNode;
}

// The mobile menu sheet, split from Nav so the Base UI dialog chunk loads on first use (Nav lazy-loads it).
export default function NavSheet({
	open,
	onOpenChange,
	sheetRef,
	finalFocus,
	title,
	closeLabel,
	children,
}: Props): ReactElement {
	return (
		<BaseDialog.Root open={open} onOpenChange={onOpenChange}>
			<BaseDialog.Portal>
				<BaseDialog.Backdrop className="fixed inset-0 z-[55] bg-backdrop/50 transition-opacity duration-200 data-ending-style:opacity-0 data-starting-style:opacity-0 motion-reduce:transition-none lg:hidden" />
				<BaseDialog.Popup
					ref={sheetRef}
					finalFocus={finalFocus}
					onClick={(e) => {
						// Following a link closes the sheet; the theme toggle keeps it open to show the change.
						const hit = (e.target as HTMLElement).closest("a,button");
						if (hit && !hit.hasAttribute("data-theme-toggle")) onOpenChange(false);
					}}
					className="fixed inset-y-0 right-0 z-[60] flex w-80 max-w-[calc(100%-3rem)] flex-col overflow-y-auto border-l border-fg bg-bg text-fg outline-none transition-transform duration-200 data-ending-style:translate-x-full data-starting-style:translate-x-full motion-reduce:transition-none lg:hidden"
				>
					<div className="flex h-14 shrink-0 items-center justify-between border-b border-border pr-2 pl-4">
						<BaseDialog.Title className="font-display text-xl font-bold">
							{title}
						</BaseDialog.Title>
						<BaseDialog.Close aria-label={closeLabel} className={iconButton}>
							<Icon name="close" />
						</BaseDialog.Close>
					</div>
					{children}
				</BaseDialog.Popup>
			</BaseDialog.Portal>
		</BaseDialog.Root>
	);
}
