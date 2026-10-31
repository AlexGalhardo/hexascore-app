import { Link } from "@inertiajs/react";
import { useT } from "../../lib/useT";
import CoverThumb from "../CoverThumb";
import type { Highlight } from "./SectionHighlights";

export default function ItemCard({ item, href }: { item: Highlight; href: string }) {
	const t = useT();

	return (
		<li className="flex items-center gap-3 py-2">
			<CoverThumb cover={item.cover} alt="" size="md" />
			<Link href={href} className="min-w-0 flex-1 truncate font-medium">
				{item.title}
			</Link>
			<span className="font-mono text-xs text-muted tabular-nums">{item.year ?? ""}</span>
			<span className="font-mono text-xs text-muted tabular-nums">
				{t("catalog.ratingCount", { count: item.ratingCount })}
			</span>
			<span
				className="w-8 text-right font-mono text-sm tabular-nums"
				title={
					item.score === null ? t("score.none") : t("score.outOf", { score: item.score })
				}
			>
				{item.score === null ? "–" : item.score}
			</span>
		</li>
	);
}
