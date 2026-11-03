import { Menu } from "@base-ui/react/menu";
import type { ComponentProps, ReactElement } from "react";

import { cn } from "../../../lib/cn";

export const DropdownMenu = Menu.Root;
export const DropdownMenuTrigger = Menu.Trigger;

export function DropdownMenuContent({
	className,
	align = "end",
	children,
}: {
	className?: string;
	align?: "start" | "center" | "end";
	children: ComponentProps<typeof Menu.Popup>["children"];
}): ReactElement {
	return (
		<Menu.Portal>
			<Menu.Positioner align={align} sideOffset={6}>
				<Menu.Popup
					className={cn(
						"z-[70] min-w-48 overflow-hidden rounded-lg border border-border bg-bg p-1 text-sm text-fg shadow-lg outline-none",
						className,
					)}
				>
					{children}
				</Menu.Popup>
			</Menu.Positioner>
		</Menu.Portal>
	);
}

export function DropdownMenuItem({
	className,
	...props
}: ComponentProps<typeof Menu.Item>): ReactElement {
	return (
		<Menu.Item
			className={cn(
				"flex min-h-9 cursor-default items-center gap-2 rounded-md px-2 py-1.5 outline-none select-none data-[highlighted]:bg-hover data-[disabled]:opacity-50 [&_svg]:size-4 [&_svg]:shrink-0",
				className,
			)}
			{...props}
		/>
	);
}

export function DropdownMenuLabel({
	className,
	...props
}: ComponentProps<typeof Menu.GroupLabel>): ReactElement {
	return (
		<Menu.GroupLabel
			className={cn("px-2 py-1.5 text-xs font-medium text-muted", className)}
			{...props}
		/>
	);
}

export function DropdownMenuSeparator({
	className,
	...props
}: ComponentProps<typeof Menu.Separator>): ReactElement {
	return <Menu.Separator className={cn("-mx-1 my-1 h-px bg-border", className)} {...props} />;
}

export const DropdownMenuGroup = Menu.Group;
