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

interface FriendRated {
	nickname: string;
	score: number;
	item: { id: string; type: Section["type"]; slug: string; title: string; cover: Cover | null };
}

// Design v1 home: one block per item type (three trending cards, then other titles), with the
// personal lists above them when present.
export default function Index({
	sections,
	trending,
	forYou,
	friends,
}: {
	sections: Section[];
	trending: Trending;
	forYou?: Partial<Record<Section["type"], ForYouList>>;
	/** Deferred (group `personal`) while the flag `social` is on. */
	friends?: FriendRated[];
}) {
	const t = useT();
	const { locale } = usePage<SharedProps>().props;

	return (
		<>
			<Seo />
			<h1 className="sr-only">{t("home.title")}</h1>
			<div className="grid gap-14">
				{friends && friends.length > 0 && (
					<section
						aria-labelledby="home-friends"
						className="grid gap-4 border-t-2 border-fg pt-4"
					>
						<h2 id="home-friends" className="text-3xl">
							{t("friendsScore.home")}
						</h2>
						<ul className="grid gap-x-6 sm:grid-cols-2 lg:grid-cols-3">
							{friends.map((f) => (
								<li
									key={`${f.item.id}-${f.nickname}`}
									className="flex items-center gap-3 border-b border-border py-2.5"
								>
									<CoverThumb cover={f.item.cover} alt="" size="md" />
									<div className="min-w-0 flex-1">
										<Link
											href={localized(
												`/${SECTION_PATHS[f.item.type]}/${f.item.slug}`,
												locale,
											)}
											className="block truncate font-display font-bold hover:text-primary"
										>
											{f.item.title}
										</Link>
										<span className="text-xs text-muted">
											{t("home.ratedBy", { name: f.nickname })}
										</span>
									</div>
									<ScoreBadge score={f.score} size="xs" />
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
			</div>
		</>
	);
}
