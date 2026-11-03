import { Link } from "@inertiajs/react";
import type { ReactElement } from "react";

import CoverThumb from "../CoverThumb";
import ScoreBadge from "../ui/ScoreBadge";
import type { Highlight } from "./SectionHighlights";

/**
 * Vertical item card for grids (landing "More …" rows, category grids): the cover, then one row of fixed
 * height so tiles never shift: title (one line, truncated) on the left, the score badge on the right.
 * The edge is a soft grey hairline (owner 2026-11-03): the category is shown by its icon, not the border.
 */
export default function ItemTile({ item, href }: { item: Highlight; href: string }): ReactElement {
	return (
		<li className="grid content-start">
			<Link
				href={href}
				className="group grid gap-2 rounded-lg border border-border bg-surface p-2 shadow-ledge-secondary hover:bg-hover outline-none focus-visible:ring-[3px] focus-visible:ring-ring/50"
			>
				<CoverThumb cover={item.cover} alt="" size="card" />
				<span
					data-tile-footer=""
					className="grid h-8 grid-cols-[1fr_auto] items-center gap-2"
				>
					<span className="truncate font-display text-sm font-bold">{item.title}</span>
					<ScoreBadge score={item.score} size="xs" />
				</span>
			</Link>
		</li>
	);
}
