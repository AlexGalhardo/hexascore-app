import { Tabs as BaseTabs } from "@base-ui/react/tabs";
import type { ComponentProps, ReactElement } from "react";

import { cn } from "../../../lib/cn";

export const Tabs = BaseTabs.Root;

export function TabsList({
	className,
	...props
}: ComponentProps<typeof BaseTabs.List>): ReactElement {
	return (
		<BaseTabs.List
			className={cn(
				"inline-flex h-9 items-center gap-1 rounded-lg bg-surface p-1",
				className,
			)}
			{...props}
		/>
	);
}

export function TabsTrigger({
	className,
	...props
}: ComponentProps<typeof BaseTabs.Tab>): ReactElement {
	return (
		<BaseTabs.Tab
			className={cn(
				"inline-flex h-7 items-center justify-center rounded-md px-3 text-sm font-medium whitespace-nowrap text-muted outline-none transition-colors focus-visible:ring-2 focus-visible:ring-primary data-[active]:bg-bg data-[active]:text-fg data-[active]:shadow-sm",
				className,
			)}
			{...props}
		/>
	);
}

export function TabsContent({
	className,
	...props
}: ComponentProps<typeof BaseTabs.Panel>): ReactElement {
	return <BaseTabs.Panel className={cn("outline-none", className)} {...props} />;
}
