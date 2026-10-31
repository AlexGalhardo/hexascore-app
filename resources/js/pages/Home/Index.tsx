import { Link, usePage } from "@inertiajs/react";
import CoverThumb, { type Cover } from "../../components/CoverThumb";
import SectionHighlights, { type Section } from "../../components/catalog/SectionHighlights";
import TrendingSection, { type TrendingList } from "../../components/catalog/TrendingSection";
import Seo from "../../components/Seo";
import { localized } from "../../lib/route";
import { useT } from "../../lib/useT";
import type { SharedProps } from "../../types";

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
}: {
	sections: Section[];
	recent: RecentRating[];
	trending: Record<Section["type"], TrendingList>;
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
				{sections.map((section) => (
					<TrendingSection
						key={`t-${section.type}`}
						type={section.type}
						list={trending[section.type]}
					/>
				))}
				{sections.map((section) => (
					<SectionHighlights key={section.type} section={section} />
				))}
			</div>
		</>
	);
}
