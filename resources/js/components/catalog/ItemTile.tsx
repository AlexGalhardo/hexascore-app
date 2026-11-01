import { Link } from "@inertiajs/react";
import type { ReactElement } from "react";

import CoverThumb from "../CoverThumb";
import ScoreBadge from "../ui/ScoreBadge";
import type { Highlight } from "./SectionHighlights";

/**
 * Vertical item card for grids (landing "More …" rows, category grids): the cover, then a footer of
 * fixed minimum height so tiles never shift: title (two lines at most) and year on the left two thirds,
 * the score badge filling the right third.
 */
export default function ItemTile({ item, href }: { item: Highlight; href: string }): ReactElement {
	return (
		<li className="grid content-start">
			<Link
				href={href}
				className="group grid gap-2 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fg"
			>
				<CoverThumb cover={item.cover} alt="" size="card" />
				<span data-tile-footer="" className="grid min-h-16 grid-cols-[2fr_1fr] gap-2">
					<span className="grid min-w-0 content-start gap-0.5">
						<span className="line-clamp-2 font-display text-sm leading-snug font-bold group-hover:text-primary">
							{item.title}
						</span>
						{item.year !== null && (
							<span className="text-xs text-muted tabular-nums">{item.year}</span>
						)}
					</span>
					<ScoreBadge score={item.score} size="fill" />
				</span>
			</Link>
		</li>
	);
}
