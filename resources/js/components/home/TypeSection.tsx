import { Link, usePage } from "@inertiajs/react";

import { localized } from "../../lib/route";
import { useT } from "../../lib/useT";
import type { SharedProps } from "../../types";
import CategoryIcon from "../CategoryIcon";
import CoverThumb from "../CoverThumb";
import ItemTile from "../catalog/ItemTile";
import { type Highlight, SECTION_PATHS, type Section } from "../catalog/SectionHighlights";
import ScoreBadge, { scoreBand } from "../ui/ScoreBadge";

export interface TrendingEntry extends Highlight {
	rank: number;
	basis: "trend" | "prior";
	/** Localized name of the item's primary genre, null when it has none. */
	genre: string | null;
}

type ItemType = Section["type"];

// Full class names so Tailwind keeps them: one accent per section (design v1).
export const ACCENT_TEXT: Record<ItemType, string> = {
	game: "text-accent-game",
	movie: "text-accent-movie",
	series: "text-accent-series",
	book: "text-accent-book",
};
const ACCENT_BORDER: Record<ItemType, string> = {
	game: "border-accent-game",
	movie: "border-accent-movie",
	series: "border-accent-series",
	book: "border-accent-book",
};

const MIN_RATINGS = 5;
const MORE = 6;

// "Other titles": most rated, then top rated, then newest, without the trending ones.
export function otherTitles(section: Section, trending: Highlight[]): Highlight[] {
	const seen = new Set(trending.map((i) => i.id));
	const out: Highlight[] = [];
	for (const item of [...section.mostRated, ...section.topRated, ...section.newest]) {
		if (seen.has(item.id)) continue;
		seen.add(item.id);
		out.push(item);
	}
	return out.slice(0, MORE);
}

// One home block per item type: three trending cards (last 7 days), then other titles.
export default function TypeSection({
	section,
	trending,
}: {
	section: Section;
	trending: TrendingEntry[];
}) {
	const t = useT();
	const { locale } = usePage<SharedProps>().props;
	const type = section.type;
	const base = localized(`/${SECTION_PATHS[type]}`, locale);
	const others = otherTitles(section, trending);
	const num = new Intl.NumberFormat(locale);
	if (trending.length === 0 && others.length === 0) return null;

	return (
		<section aria-labelledby={`home-${type}`} className="grid gap-5 border-t-2 border-fg pt-4">
			{/* Three zones (owner round 2026-11-01 item 23): 1fr auto 1fr keeps the count centred whatever
			    the name and link widths; below sm the zones stack. */}
			<div
				data-section-header=""
				className="grid min-w-0 gap-y-1 sm:grid-cols-[1fr_auto_1fr] sm:items-center sm:gap-x-4"
			>
				<div className="flex min-w-0 items-center gap-3">
					<CategoryIcon type={type} className={`size-7 ${ACCENT_TEXT[type]}`} />
					<h2 id={`home-${type}`} className="text-3xl">
						<Link href={base}>{t(`catalog.title.${type}`)}</Link>
					</h2>
				</div>
				<span className="text-xl font-semibold text-muted tabular-nums sm:text-center">
					{t("home.titles", { count: section.count, n: num.format(section.count) })}
				</span>
				<Link
					href={base}
					className="text-sm font-semibold text-primary underline-offset-2 hover:underline sm:justify-self-end sm:text-right"
				>
					{t(`home.seeAll.${type}`)} <span aria-hidden="true">→</span>
				</Link>
			</div>

			{trending.length > 0 && (
				<ol className="grid gap-4 md:grid-cols-3 md:gap-5">
					{trending.slice(0, 3).map((item) => (
						<li
							key={item.id}
							className={`flex gap-4 border-t-2 bg-surface p-3 ${ACCENT_BORDER[type]}`}
						>
							<CoverThumb cover={item.cover} alt="" size="trend" />
							<div className="grid min-w-0 content-start gap-1">
								<p className={`text-xs font-semibold ${ACCENT_TEXT[type]}`}>
									{t("home.trendingRank", { rank: item.rank })}
								</p>
								<Link
									href={`${base}/${item.slug}?from=trending`}
									className="font-display text-xl leading-tight font-bold hover:text-primary"
								>
									{item.title}
								</Link>
								<p className="text-xs text-muted">
									{[item.year, item.genre].filter(Boolean).join(" · ")}
								</p>
								<div className="mt-2 flex items-center gap-3">
									<ScoreBadge score={item.score} size="sm" />
									<p className="text-xs">
										<span className="block font-semibold">
											{item.score === null
												? t("item.notEnough")
												: t(`item.band.${scoreBand(item.score)}`)}
										</span>
										<span className="text-muted">
											{item.score === null
												? t("home.ratingsOfMin", {
														count: item.ratingCount,
														min: MIN_RATINGS,
													})
												: t("home.fromRatings", {
														count: item.ratingCount,
														n: num.format(item.ratingCount),
													})}
										</span>
									</p>
								</div>
							</div>
						</li>
					))}
				</ol>
			)}

			{others.length > 0 && (
				<div className="grid gap-3">
					<h3 className="text-xs font-semibold tracking-widest text-muted uppercase">
						{t(`home.more.${type}`)}
					</h3>
					<ul className="grid grid-cols-3 gap-x-4 gap-y-6 md:grid-cols-6 md:gap-x-5">
						{others.map((item) => (
							<ItemTile key={item.id} item={item} href={`${base}/${item.slug}`} />
						))}
					</ul>
				</div>
			)}
		</section>
	);
}
