import type { ReactNode } from "react";
import { NavIcon } from "@/components/layout/nav-icons";
import type { ItemType } from "@/lib/db/schema/catalog";
import { SECTION_ACCENT, SECTION_PATHS } from "@/lib/sections";

const ICON = { sm: "size-3.5", md: "size-4", lg: "size-5", xl: "size-6" } as const;

type Props = {
	type: ItemType;
	/** The visible text (the caller picks singular, plural or a count); the icon is decorative. */
	children: ReactNode;
	size?: keyof typeof ICON;
	className?: string;
};

/** Every category name in the UI: the section icon in the section accent, then the label (docs/frontend.md). */
export function CategoryLabel({ type, children, size = "sm", className = "" }: Props) {
	return (
		<span className={`inline-flex items-center gap-1.5 ${className}`}>
			<NavIcon name={SECTION_PATHS[type]} className={`${ICON[size]} ${SECTION_ACCENT[type].text}`} />
			{children}
		</span>
	);
}
