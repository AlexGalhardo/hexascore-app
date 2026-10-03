import { SECTION_ACCENT } from "@/lib/sections";
import type { ListType } from "@/lib/social/constants";

/** The list's category: a section-accent dot (decorative) plus the translated type label. */
export function ListTypeBadge({ type, label }: { type: ListType; label: string }) {
	return (
		<span className="inline-flex items-center gap-1.5 text-xs font-medium">
			<span aria-hidden="true" className={`size-2 shrink-0 rounded-full ${SECTION_ACCENT[type].bg}`} />
			{label}
		</span>
	);
}
