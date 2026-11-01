import { Link, usePage } from "@inertiajs/react";
import CoverThumb, { type Cover } from "../../components/CoverThumb";
import ForYouListView, { type ForYouList } from "../../components/catalog/ForYouList";
import { SECTION_PATHS, type Section } from "../../components/catalog/SectionHighlights";
import TypeSection, { type TrendingEntry } from "../../components/home/TypeSection";
import Seo from "../../components/Seo";
import ScoreBadge from "../../components/ui/ScoreBadge";
import { localized } from "../../lib/route";
import { useT } from "../../lib/useT";
import type { SharedProps } from "../../types";

type Trending = Partial<Record<Section["type"], { computedAt: string; items: TrendingEntry[] }>>;

interface RecentRating {
	id: string;
	type: Section["type"];
	slug: string;
	title: string;
	score: number;
	by: string;
	cover: Cover | null;
}

// Design v1 home: one block per item type (three trending cards, then other titles), with the
// members' recent ratings and the personal lists above them when present.
export default function Index({
	sections,
	recent,
	trending,
	forYou,
}: {
	sections: Section[];
	recent: RecentRating[];
	trending: Trending;
	forYou?: Partial<Record<Section["type"], ForYouList>>;
}) {
	const t = useT();
	const { locale } = usePage<SharedProps>().props;

	return (
		<>
			<Seo />
			<h1 className="sr-only">{t("home.title")}</h1>
			<div className="grid gap-14">
				{recent.length > 0 && (
					<section
						aria-labelledby="home-recent"
						className="grid gap-4 border-t-2 border-fg pt-4"
					>
						<h2 id="home-recent" className="text-3xl">
							{t("home.recent")}
						</h2>
						<ul className="grid gap-x-6 sm:grid-cols-2 lg:grid-cols-3">
							{recent.map((r) => (
								<li
									key={`${r.id}-${r.by}`}
									className="flex items-center gap-3 border-b border-border py-2.5"
								>
									<CoverThumb cover={r.cover} alt="" size="md" />
									<div className="min-w-0 flex-1">
										<Link
											href={localized(
												`/${SECTION_PATHS[r.type]}/${r.slug}`,
												locale,
											)}
											className="block truncate font-display font-bold hover:text-primary"
										>
											{r.title}
										</Link>
										<span className="text-xs text-muted">
											{t("home.ratedBy", { name: r.by })}
										</span>
									</div>
									<ScoreBadge score={r.score} size="xs" />
								</li>
							))}
						</ul>
					</section>
				)}
				{Object.entries(forYou ?? {}).map(([type, list]) => (
					<section
						key={`foryou-${type}`}
						aria-labelledby={`home-foryou-${type}`}
						className="grid gap-4 border-t-2 border-fg pt-4"
					>
						<h2 id={`home-foryou-${type}`} className="text-3xl">
							{t("forYou.home")}: {t(`catalog.title.${type}`)}
						</h2>
						<ForYouListView type={type as Section["type"]} list={list} />
					</section>
				))}
				{sections.map((section) => (
					<TypeSection
						key={section.type}
						section={section}
						trending={trending[section.type]?.items ?? []}
					/>
				))}
				<p className="text-xs text-muted">{t("home.trendingNote")}</p>
			</div>
		</>
	);
}
