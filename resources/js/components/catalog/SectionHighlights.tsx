import { Link, usePage } from "@inertiajs/react";
import { localized } from "../../lib/route";
import { useT } from "../../lib/useT";
import type { SharedProps } from "../../types";
import CoverThumb, { type Cover } from "../CoverThumb";

export interface Highlight {
	id: string;
	type: string;
	slug: string;
	title: string;
	year: number | null;
	score: number | null;
	ratingCount: number;
	cover: Cover | null;
}

export interface Section {
	type: "game" | "movie" | "series" | "book";
	count: number;
	topRated: Highlight[];
	newest: Highlight[];
	mostRated: Highlight[];
}

export const SECTION_PATHS: Record<Section["type"], string> = {
	game: "games",
	movie: "movies",
	series: "series",
	book: "books",
};

const BLOCKS = [
	{ key: "topRated", sort: "-score" },
	{ key: "newest", sort: "-release_date" },
	{ key: "mostRated", sort: "-rating_count" },
] as const;

// Landing block for one item type: three short ranked lists, stacked on mobile.
export default function SectionHighlights({ section }: { section: Section }) {
	const t = useT();
	const { locale } = usePage<SharedProps>().props;
	const path = localized(`/${SECTION_PATHS[section.type]}`, locale);
	const title = t(`catalog.title.${section.type}`);

	return (
		<section aria-labelledby={`home-${section.type}`} className="grid gap-3">
			<div className="flex items-baseline justify-between gap-3 border-b-2 border-border pb-1">
				<h2 id={`home-${section.type}`} className="text-2xl font-bold">
					<Link href={path}>{title}</Link>
				</h2>
				<Link href={path} className="text-sm font-medium">
					{t("catalog.landing.seeAll", { section: title })} →
				</Link>
			</div>
			<div className="grid gap-6 md:grid-cols-3">
				{BLOCKS.map((block) => {
					const rows = section[block.key];
					return (
						<div key={block.key}>
							<h3 className="mb-1 text-xs font-semibold uppercase tracking-wide text-muted">
								<Link href={`${path}?sort=${block.sort}`}>
									{t(`catalog.landing.${block.key}`)}
								</Link>
							</h3>
							{rows.length ? (
								<ol className="divide-y divide-border">
									{rows.map((item, rank) => (
										<li
											key={item.id}
											className="flex items-center gap-2 py-1.5"
										>
											<span className="w-5 shrink-0 font-mono text-xs text-muted tabular-nums">
												{rank + 1}
											</span>
											<CoverThumb cover={item.cover} alt="" />
											<Link
												href={`${path}/${item.slug}`}
												className="min-w-0 flex-1 truncate text-sm"
											>
												{item.title}
											</Link>
											<span className="font-mono text-xs text-muted tabular-nums">
												{block.key === "mostRated"
													? item.ratingCount
													: item.year}
											</span>
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
							) : (
								<p className="py-2 text-sm text-muted">
									{t("catalog.landing.empty")}
								</p>
							)}
						</div>
					);
				})}
			</div>
		</section>
	);
}
