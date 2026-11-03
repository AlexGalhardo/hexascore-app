import { Link } from "@inertiajs/react";
import type { ReactElement, ReactNode } from "react";

import CoverThumb, { type CoverSize } from "../CoverThumb";
import ScoreBadge from "../ui/ScoreBadge";
import type { Highlight } from "./SectionHighlights";

/**
 * Vertical item card for grids (landing "More …" rows, category grids): the cover, then one row of fixed
 * height so tiles never shift: title (one line, truncated) on the left, the score badge on the right
 * (or the given badge, e.g. a release date on upcoming cards).
 * The edge is a soft grey hairline (owner 2026-11-03): the category is shown by its icon, not the border.
 */
export default function ItemTile({
	item,
	href,
	badge,
	footer,
	coverSize = "card",
}: {
	item: Highlight;
	href: string;
	/** null hides the badge; undefined shows the score. */
	badge?: ReactNode | null;
	/** Replaces the item title in the bottom strip. */
	footer?: ReactNode;
	coverSize?: CoverSize;
}): ReactElement {
	return (
		<li className="grid content-start">
			<Link
				href={href}
				className="group relative block overflow-hidden rounded-sm border border-border bg-surface hover:border-fg focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fg"
			>
				<CoverThumb cover={item.cover} alt="" size={coverSize} />
				{badge !== null && (
					<span className="absolute top-2 right-2">
						{badge ?? <ScoreBadge score={item.score} size="xs" />}
					</span>
				)}
				<span className="absolute inset-x-0 bottom-0 truncate bg-gradient-to-t from-black/85 to-transparent p-2 pt-7 font-display text-sm font-bold text-white">
					{footer ?? item.title}
				</span>
			</Link>
		</li>
	);
}
