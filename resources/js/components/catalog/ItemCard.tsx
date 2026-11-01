import { Link } from "@inertiajs/react";
import { useT } from "../../lib/useT";
import CoverThumb from "../CoverThumb";
import ScoreBadge from "../ui/ScoreBadge";
import type { Highlight } from "./SectionHighlights";

export default function ItemCard({ item, href }: { item: Highlight; href: string }) {
	const t = useT();
	const meta = [item.year, t("catalog.ratingCount", { count: item.ratingCount })]
		.filter((v) => v !== null && v !== "")
		.join(" · ");

	return (
		<li className="flex items-center gap-4 border-b border-border py-3">
			<CoverThumb cover={item.cover} alt="" size="md" />
			<div className="min-w-0 flex-1">
				<Link
					href={href}
					className="block truncate font-display text-lg leading-snug font-bold hover:text-primary focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fg"
				>
					{item.title}
				</Link>
				<p className="text-[13px] text-muted tabular-nums">{meta}</p>
			</div>
			<ScoreBadge score={item.score} />
		</li>
	);
}
