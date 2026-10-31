import { Link, usePage } from "@inertiajs/react";
import { localized } from "../../lib/route";
import { useT } from "../../lib/useT";
import type { SharedProps } from "../../types";
import CoverThumb from "../CoverThumb";
import { type Highlight, SECTION_PATHS, type Section } from "./SectionHighlights";

export interface TrendingEntry extends Highlight {
	rank: number;
	basis: "trend" | "prior";
}

export interface TrendingList {
	computedAt: string | null;
	items: TrendingEntry[];
}

// Top 10 of one type over the last 30 days. A type without entries renders nothing (SDD 08 §6).
export default function TrendingSection({
	type,
	list,
}: {
	type: Section["type"];
	list?: TrendingList;
}) {
	const t = useT();
	const { locale } = usePage<SharedProps>().props;
	// The prop is absent off Postgres (the home controller skips the snapshot read there).
	if (!list || list.items.length === 0) return null;

	const path = localized(`/${SECTION_PATHS[type]}`, locale);
	const typeTitle = t(`catalog.title.${type}`);
	const allPrior = list.items.every((item) => item.basis === "prior");

	return (
		<section aria-labelledby={`trending-${type}`} className="grid gap-3">
			<h2
				id={`trending-${type}`}
				className="border-b-2 border-border pb-1 text-2xl font-bold"
			>
				{allPrior
					? t("home.trending.popularType", { type: typeTitle })
					: t("home.trending.title", { type: typeTitle })}
			</h2>
			<ol className="grid gap-x-6 sm:grid-cols-2 sm:[&>li:nth-child(n)]:border-t-0">
				{list.items.map((item) => (
					<li
						key={item.id}
						className="flex items-center gap-2 border-b border-border py-1.5"
					>
						<span className="w-5 shrink-0 font-mono text-xs text-muted tabular-nums">
							{item.rank}
						</span>
						<CoverThumb cover={item.cover} alt="" />
						<Link
							href={`${path}/${item.slug}?from=trending`}
							className="min-w-0 flex-1 truncate text-sm"
						>
							{item.title}
						</Link>
						{!allPrior && item.basis === "prior" && (
							<span className="rounded border border-border px-1.5 text-xs text-muted">
								{t("home.trending.popular")}
							</span>
						)}
						<span
							className="font-mono text-xs tabular-nums"
							title={
								item.score === null
									? t("score.none")
									: t("score.outOf", { score: item.score })
							}
						>
							{item.score === null ? "–" : item.score}
						</span>
					</li>
				))}
			</ol>
		</section>
	);
}
