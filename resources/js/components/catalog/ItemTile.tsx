import { Link } from "@inertiajs/react";
import type { ReactElement } from "react";

import CoverThumb from "../CoverThumb";
import ScoreBadge from "../ui/ScoreBadge";
import type { Highlight } from "./SectionHighlights";

// Full class names so Tailwind keeps them: one border colour per category (theme tokens in app.css).
export const CATEGORY_BORDER: Record<Highlight["type"], string> = {
	game: "border-cat-game",
	movie: "border-cat-movie",
	series: "border-cat-series",
	book: "border-cat-book",
};

/**
 * Vertical item card for grids (landing "More …" rows, category grids), bordered in its category colour:
 * the cover, then one row of fixed height so tiles never shift: title (one line, truncated) on the left,
 * the score badge on the right.
 */
export default function ItemTile({ item, href }: { item: Highlight; href: string }): ReactElement {
	return (
		<li className="grid content-start">
			<Link
				href={href}
				className={`group grid gap-2 border-2 bg-surface p-2 shadow-ledge-secondary hover:bg-fg hover:text-bg focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fg ${CATEGORY_BORDER[item.type]}`}
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
