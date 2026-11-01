import { Link, usePage } from "@inertiajs/react";

import { localized } from "../../lib/route";
import { useT } from "../../lib/useT";
import type { SharedProps } from "../../types";
import CoverThumb from "../CoverThumb";
import { type Highlight, SECTION_PATHS, type Section } from "../catalog/SectionHighlights";
import ScoreBadge, { scoreBand } from "../ui/ScoreBadge";

export interface TrendingEntry extends Highlight {
	rank: number;
	basis: "trend" | "prior";
}

type ItemType = Section["type"];

// Full class names so Tailwind keeps them: one accent per section (design v1).
const ACCENT_BG: Record<ItemType, string> = {
	game: "bg-accent-game",
	movie: "bg-accent-movie",
	series: "bg-accent-series",
	book: "bg-accent-book",
};
const ACCENT_TEXT: Record<ItemType, string> = {
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
			<div className="flex flex-wrap items-baseline justify-between gap-x-4 gap-y-1">
				<div className="flex items-baseline gap-3">
					<span
						aria-hidden="true"
						className={`size-2.5 self-center ${ACCENT_BG[type]}`}
					/>
					<h2 id={`home-${type}`} className="text-3xl">
						<Link href={base}>{t(`catalog.title.${type}`)}</Link>
					</h2>
					<span className="text-xs text-muted tabular-nums">
						{t("home.titles", { count: section.count, n: num.format(section.count) })}
					</span>
				</div>
				<Link
					href={base}
					className="text-sm font-semibold text-primary underline-offset-2 hover:underline"
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
									{item.basis === "prior"
										? t("home.popularRank", { rank: item.rank })
										: t("home.trendingRank", { rank: item.rank })}
								</p>
								<Link
									href={`${base}/${item.slug}?from=trending`}
									className="font-display text-xl leading-tight font-bold hover:text-primary"
								>
									{item.title}
								</Link>
								<p className="text-xs text-muted">
									{[item.year, t(`itemTypes.${type}`)]
										.filter(Boolean)
										.join(" · ")}
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
							<li key={item.id} className="grid content-start gap-2">
								<Link
									href={`${base}/${item.slug}`}
									className="group grid gap-2 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fg"
								>
									<CoverThumb cover={item.cover} alt="" size="card" />
									<span className="font-display text-sm leading-snug font-bold group-hover:text-primary">
										{item.title}
									</span>
								</Link>
								<div className="flex items-center justify-between gap-2">
									<span className="text-xs text-muted tabular-nums">
										{item.year}
									</span>
									<ScoreBadge score={item.score} size="xs" />
								</div>
							</li>
						))}
					</ul>
				</div>
			)}
		</section>
	);
}
