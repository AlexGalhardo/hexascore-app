import { Link, usePage } from "@inertiajs/react";
import CoverThumb, { type Cover } from "../../components/CoverThumb";
import ForYouListView, { type ForYouList } from "../../components/catalog/ForYouList";
import SectionHighlights, {
	type Highlight,
	type Section,
} from "../../components/catalog/SectionHighlights";
import Seo from "../../components/Seo";
import { localized } from "../../lib/route";
import { useT } from "../../lib/useT";
import type { SharedProps } from "../../types";

interface TrendingEntry extends Highlight {
	rank: number;
	basis: "trend" | "prior";
}

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

const PATHS: Record<Section["type"], string> = {
	game: "games",
	movie: "movies",
	series: "series",
	book: "books",
};

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
			<div className="grid gap-12">
				<header className="grid justify-items-start gap-4 border-b-2 border-border pb-8 pt-4">
					<h1 className="max-w-3xl text-3xl font-bold leading-tight sm:text-4xl">
						{t("home.title")}
					</h1>
					<Link
						href={localized("/search", locale)}
						className="rounded-md bg-indigo-600 px-5 py-2.5 font-semibold text-white hover:bg-indigo-500"
					>
						{t("home.explore")}
					</Link>
				</header>
				{recent.length > 0 && (
					<section aria-labelledby="home-recent" className="grid gap-3">
						<h2 id="home-recent" className="text-2xl font-bold">
							{t("home.recent")}
						</h2>
						<ul className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
							{recent.map((r) => (
								<li key={`${r.id}-${r.by}`} className="flex items-center gap-3">
									<CoverThumb cover={r.cover} alt="" size="md" />
									<div className="min-w-0 flex-1">
										<Link
											href={localized(`/${PATHS[r.type]}/${r.slug}`, locale)}
											className="block truncate font-medium"
										>
											{r.title}
										</Link>
										<span className="text-xs text-muted">
											{t("home.ratedBy", { name: r.by })}
										</span>
									</div>
									<span
										className="font-mono text-sm tabular-nums"
										title={t("score.outOf", { score: r.score })}
									>
										{r.score}
									</span>
								</li>
							))}
						</ul>
					</section>
				)}
				{(Object.keys(PATHS) as Section["type"][]).map((type) => {
					const items = trending[type]?.items ?? [];
					if (items.length === 0) return null;
					const allPrior = items.every((i) => i.basis === "prior");
					const label = t(`catalog.title.${type}`);
					return (
						<section
							key={`trending-${type}`}
							aria-labelledby={`home-trending-${type}`}
							className="grid gap-3"
						>
							<h2 id={`home-trending-${type}`} className="text-2xl font-bold">
								{allPrior
									? t("home.popularType", { type: label })
									: `${t("home.trendingWeek")}: ${label}`}
							</h2>
							<ol className="grid gap-3 sm:grid-cols-3">
								{items.map((item) => (
									<li key={item.id} className="flex items-center gap-3">
										<span className="w-5 font-mono text-sm text-muted tabular-nums">
											{item.rank}
										</span>
										<CoverThumb cover={item.cover} alt="" size="md" />
										<div className="min-w-0 flex-1">
											<Link
												href={localized(
													`/${PATHS[type]}/${item.slug}?from=trending`,
													locale,
												)}
												className="block truncate font-medium"
											>
												{item.title}
											</Link>
											{item.basis === "prior" && !allPrior && (
												<span className="text-xs text-muted">
													{t("home.popular")}
												</span>
											)}
										</div>
									</li>
								))}
							</ol>
						</section>
					);
				})}
				{Object.entries(forYou ?? {}).map(([type, list]) => (
					<section
						key={`foryou-${type}`}
						aria-labelledby={`home-foryou-${type}`}
						className="grid gap-3"
					>
						<h2 id={`home-foryou-${type}`} className="text-2xl font-bold">
							{t("forYou.home")}: {t(`catalog.title.${type}`)}
						</h2>
						<ForYouListView type={type as Section["type"]} list={list} />
					</section>
				))}
				{sections.map((section) => (
					<SectionHighlights key={section.type} section={section} />
				))}
			</div>
		</>
	);
}
