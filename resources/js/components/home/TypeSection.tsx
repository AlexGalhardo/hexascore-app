import { Link, usePage } from "@inertiajs/react";
import { type ReactElement, useEffect, useMemo, useState } from "react";

import { localized } from "../../lib/route";
import { useT } from "../../lib/useT";
import type { SharedProps } from "../../types";
import CategoryIcon from "../CategoryIcon";
import CoverThumb from "../CoverThumb";
import ItemTile from "../catalog/ItemTile";
import { type Highlight, SECTION_PATHS, type Section } from "../catalog/SectionHighlights";
import ScoreBadge, { scoreBand } from "../ui/ScoreBadge";
import Carousel from "./Carousel";

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
const PER_SLIDE = 3;
const MORE = 12;

/** n distinct items of pool in random order (partial Fisher-Yates); rng is injectable for tests. */
export function pickRandom<T>(pool: readonly T[], n: number, rng: () => number = Math.random): T[] {
	const a = [...pool];
	const k = Math.min(n, a.length);
	for (let i = 0; i < k; i++) {
		const j = i + Math.floor(rng() * (a.length - i));
		[a[i], a[j]] = [a[j], a[i]];
	}
	return a.slice(0, k);
}

// One home block per item type (owner round 2026-11-01 item 26): a carousel of the 12 trending items of the last
// 30 days, 3 per slide, then 12 "More …" tiles. The server sends a pool and the page picks 12 at random after
// hydration: the anonymous landing HTML is cached, so the first paint shows the pool's first 12 and only then
// reshuffles into the same fixed grid (no layout shift).
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
	const candidates = useMemo(() => {
		const shown = new Set(trending.map((i) => i.id));
		return section.pool.filter((i) => !shown.has(i.id));
	}, [section.pool, trending]);
	const [others, setOthers] = useState(() => candidates.slice(0, MORE));
	useEffect(() => setOthers(pickRandom(candidates, MORE)), [candidates]);
	const num = new Intl.NumberFormat(locale);
	if (trending.length === 0 && others.length === 0) return null;

	return (
		<section aria-labelledby={`home-${type}`} className="grid gap-5">
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
				<Carousel
					label={t("home.carousel.label", { type: t(`catalog.title.${type}`) })}
					slides={Array.from(
						{ length: Math.ceil(trending.length / PER_SLIDE) },
						(_, n) => ({
							key: trending[n * PER_SLIDE].id,
							node: (
								<ol
									start={n * PER_SLIDE + 1}
									className="grid gap-4 md:grid-cols-3 md:gap-5"
								>
									{trending
										.slice(n * PER_SLIDE, (n + 1) * PER_SLIDE)
										.map((item) => (
											<TrendCard
												key={item.id}
												item={item}
												type={type}
												base={base}
												eager={n === 0}
											/>
										))}
								</ol>
							),
						}),
					)}
				/>
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

function TrendCard({
	item,
	type,
	base,
	eager,
}: {
	item: TrendingEntry;
	type: ItemType;
	base: string;
	/** Only the first slide's covers load eagerly (the LCP candidate). */
	eager: boolean;
}): ReactElement {
	const t = useT();
	const { locale } = usePage<SharedProps>().props;
	const num = new Intl.NumberFormat(locale);
	return (
		<li key={item.id} className={`flex gap-4 border-t-2 bg-surface p-3 ${ACCENT_BORDER[type]}`}>
			<CoverThumb cover={item.cover} alt="" size="trend" priority={eager} />
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
	);
}
