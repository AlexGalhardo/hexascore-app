import { Link, usePage } from "@inertiajs/react";

import { localized } from "../../lib/route";
import { useT } from "../../lib/useT";
import type { SharedProps } from "../../types";
import ItemTile from "../catalog/ItemTile";
import { type Highlight, SECTION_PATHS, type Section } from "../catalog/SectionHighlights";
import Carousel from "./Carousel";

export interface TrendingEntry extends Highlight {
	rank: number;
	basis: "trend" | "prior" | "pinned";
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

const PER_SLIDE = 4;
const MORE = 12;

// One block per type: upcoming releases in the carousel and released 30-day trending in the grid.
export default function TypeSection({
	section,
	upcoming = [],
}: {
	section: Section;
	upcoming?: Highlight[];
}) {
	const t = useT();
	const { locale } = usePage<SharedProps>().props;
	const type = section.type;
	const base = localized(`/${SECTION_PATHS[type]}`, locale);
	const num = new Intl.NumberFormat(locale);
	const others = section.pool.slice(0, MORE);
	if (upcoming.length === 0 && others.length === 0) return null;

	return (
		<section aria-labelledby={`home-${type}`} className="grid gap-5">
			<div
				data-section-header=""
				className="flex min-w-0 flex-wrap items-center justify-between gap-3"
			>
				<div className="flex min-w-0 items-center gap-3">
					<h2
						id={`home-${type}`}
						className="text-xs font-semibold tracking-widest text-muted uppercase"
					>
						<Link href={base}>
							{t("home.mostWaited", { type: t(`catalog.title.${type}`) })}
						</Link>
					</h2>
				</div>
				<Link
					href={base}
					className="text-sm font-semibold text-primary underline-offset-2 hover:underline sm:justify-self-end sm:text-right"
				>
					{t(`home.seeAll.${type}`, {
						count: section.count,
						n: num.format(section.count),
					})}
				</Link>
			</div>

			{upcoming.length > 0 && (
				<Carousel
					label={t("home.mostWaited", { type: t(`catalog.title.${type}`) })}
					slides={Array.from(
						{ length: Math.ceil(upcoming.slice(0, MORE).length / PER_SLIDE) },
						(_, n) => ({
							key: upcoming[n * PER_SLIDE].id,
							node: (
								<ol
									start={n * PER_SLIDE + 1}
									className="grid grid-cols-2 gap-x-4 gap-y-6 md:grid-cols-4 md:gap-x-6 lg:gap-x-8"
								>
									{upcoming
										.slice(n * PER_SLIDE, Math.min((n + 1) * PER_SLIDE, MORE))
										.map((item) => (
											<ItemTile
												key={item.id}
												item={item}
												href={`${base}/${item.slug}`}
												coverSize="compact"
												badge={
													<span className="rounded-sm bg-[#ff7a1a] px-1.5 py-0.5 text-xs font-bold text-black">
														{formatUpcomingRelease(item, locale)}
													</span>
												}
											/>
										))}
								</ol>
							),
						}),
					)}
				/>
			)}
			{upcoming.length === 0 && (
				<p className="rounded-sm border border-border bg-surface p-5 text-sm text-muted">
					{t("home.noUpcoming")}
				</p>
			)}

			{others.length > 0 && (
				<div className="grid gap-3">
					<h3 className="text-xs font-semibold tracking-widest text-muted uppercase">
						{t("home.trending30Days", { type: t(`catalog.title.${type}`) })}
					</h3>
					<ul className="grid grid-cols-2 gap-x-4 gap-y-6 sm:grid-cols-3 md:grid-cols-6 md:gap-x-5">
						{others.map((item) => (
							<ItemTile key={item.id} item={item} href={`${base}/${item.slug}`} />
						))}
					</ul>
				</div>
			)}
		</section>
	);
}

function formatUpcomingRelease(item: Highlight, locale: string): string {
	if (!item.releaseDate || item.releasePrecision === "unknown") return String(item.year ?? "");
	const date = new Date(`${item.releaseDate}T00:00:00Z`);
	if (Number.isNaN(date.getTime())) return String(item.year ?? "");
	if (item.releasePrecision === "year") return String(date.getUTCFullYear());
	const month = new Intl.DateTimeFormat(locale, { month: "long", timeZone: "UTC" }).format(date);
	const namedMonth = month[0].toLocaleUpperCase(locale) + month.slice(1);
	const year = date.getUTCFullYear();
	if (item.releasePrecision === "month") return `${namedMonth} ${year}`;
	const day = new Intl.DateTimeFormat(locale, { day: "2-digit", timeZone: "UTC" }).format(date);
	return `${day} ${namedMonth} ${year}`;
}
