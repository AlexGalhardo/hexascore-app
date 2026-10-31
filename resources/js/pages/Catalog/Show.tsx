import { Link, usePage } from "@inertiajs/react";
import CoverThumb, { type Cover } from "../../components/CoverThumb";
import AddToList, { type ListChoice } from "../../components/item/AddToList";
import Editions from "../../components/item/Editions";
import RatingPanel from "../../components/item/RatingPanel";
import StatusPicker from "../../components/item/StatusPicker";
import TimeToBeat, { type Medians } from "../../components/item/TimeToBeat";
import Trailers from "../../components/item/Trailers";
import type { EditionRow, ItemType, Me, RatingSummary } from "../../components/item/types";
import Seo, { truncate } from "../../components/Seo";
import { localized } from "../../lib/route";
import { useT } from "../../lib/useT";
import type { SharedProps } from "../../types";

const SECTIONS: Record<ItemType, string> = {
	game: "games",
	movie: "movies",
	series: "series",
	book: "books",
};

interface Item {
	id: string;
	type: ItemType;
	slug: string;
	title?: string;
	synopsis?: { text: string; licence: string; sourceUrl?: string };
	releaseDate?: string;
	runtimeMinutes?: number;
	pageCount?: number;
}

interface EntityLink {
	id: string;
	kind: string;
	slug: string;
	code?: string;
	name: string;
	character?: string;
}

interface Props {
	item: Item;
	ratingSummary: RatingSummary;
	entities: Record<string, EntityLink[]>;
	editions: EditionRow[] | null;
	editionCount: number;
	timeToBeat: Medians | null;
	me?: Me;
	myLists?: ListChoice[];
	trailers?: { youtubeId: string; title?: string }[];
	cover?: Cover | null;
}

const LD_TYPE: Record<ItemType, string> = {
	game: "VideoGame",
	movie: "Movie",
	series: "TVSeries",
	book: "Book",
};

const ENTITY_PATH: Record<string, string> = {
	person: "people",
	organization: "companies",
	franchise: "franchises",
	platform: "platforms",
	genre: "genres",
	game_mode: "game-modes",
	country: "countries",
	language: "languages",
};

const ROLE_ORDER = [
	"director",
	"writer",
	"creator",
	"author",
	"developer",
	"publisher",
	"network",
	"production_company",
	"genre",
	"platform",
	"game_mode",
	"franchise",
	"series",
	"cast",
];

export default function Show({
	item,
	ratingSummary,
	entities,
	editions,
	editionCount,
	timeToBeat,
	me,
	myLists,
	trailers = [],
	cover = null,
}: Props) {
	const t = useT();
	const { locale, auth } = usePage<SharedProps>().props;
	const title = item.title ?? item.slug;
	const loggedIn = Boolean(auth?.user);
	const roles = ROLE_ORDER.filter((r) => entities[r]?.length);

	const jsonLd: object = {
		"@context": "https://schema.org",
		"@type": LD_TYPE[item.type],
		name: title,
		...(item.releaseDate && { datePublished: item.releaseDate }),
		...(ratingSummary.score !== null && {
			aggregateRating: {
				"@type": "AggregateRating",
				ratingValue: ratingSummary.score,
				bestRating: 10,
				worstRating: 0.5,
				ratingCount: ratingSummary.ratingCount,
			},
		}),
	};

	return (
		<>
			<Seo
				title={title}
				description={truncate(
					item.synopsis?.text ??
						t("item.metaDescription", { title, type: t(`itemTypes.${item.type}`) }),
					160,
				)}
				jsonLd={[jsonLd]}
			/>
			<article className="grid gap-8">
				<header className="flex items-start gap-4">
					<CoverThumb cover={cover} alt={title} size="lg" priority />
					<div className="grid min-w-0 gap-2">
						<h1 className="text-3xl font-bold">{title}</h1>
						<Link
							href={localized(`/suggest/item/${item.id}`, locale)}
							className="text-sm underline"
						>
							{t("community.suggest.link")}
						</Link>
						<Link
							href={localized(`/${SECTIONS[item.type]}/${item.slug}/changes`, locale)}
							className="text-sm underline"
						>
							{t("community.recent.link")}
						</Link>
						<p className="text-muted">
							{t(`itemTypes.${item.type}`)}
							{item.releaseDate && ` · ${item.releaseDate.slice(0, 4)}`}
						</p>
						<p
							data-testid="score-badge"
							className="inline-flex w-fit items-center rounded-md bg-surface px-3 py-1 text-lg font-bold"
						>
							{ratingSummary.score === null
								? t("item.notEnough")
								: t("score.outOf", { score: ratingSummary.score })}
						</p>
					</div>
				</header>

				<div className="grid gap-8 lg:grid-cols-[minmax(0,1fr)_24rem]">
					<div className="order-1 lg:order-2 lg:sticky lg:top-4 lg:self-start">
						<RatingPanel
							key={item.id + (me?.rating ? "r" : "n")}
							itemId={item.id}
							type={item.type}
							summary={ratingSummary}
							mine={me?.rating}
							canWrite={loggedIn}
						/>
					</div>
					<div className="order-2 grid min-w-0 gap-8 lg:order-1">
						<div className="grid gap-8 md:grid-cols-[3fr_2fr]">
							<section aria-labelledby="synopsis-heading" className="grid gap-2">
								<h2 id="synopsis-heading" className="text-xl font-bold">
									{t("item.synopsis")}
								</h2>
								{item.synopsis ? (
									<>
										<p>{item.synopsis.text}</p>
										<p className="text-sm text-muted">
											{t("item.license", { license: item.synopsis.licence })}
											{item.synopsis.sourceUrl && (
												<>
													{" · "}
													<a
														href={item.synopsis.sourceUrl}
														rel="noopener noreferrer nofollow"
													>
														{t("item.source")}
													</a>
												</>
											)}
										</p>
									</>
								) : (
									<p>{t("item.noSynopsis")}</p>
								)}
							</section>

							<section aria-labelledby="details-heading" className="grid gap-2">
								<h2 id="details-heading" className="text-xl font-bold">
									{t("item.details")}
								</h2>
								<dl className="grid grid-cols-[max-content_1fr] gap-x-4 gap-y-1">
									{item.runtimeMinutes !== undefined && (
										<>
											<dt>{t("item.fields.runtime")}</dt>
											<dd>
												{t("item.runtimeValue", {
													minutes: item.runtimeMinutes,
												})}
											</dd>
										</>
									)}
									{item.pageCount !== undefined && (
										<>
											<dt>{t("item.fields.pages")}</dt>
											<dd>{item.pageCount}</dd>
										</>
									)}
									{roles.map((role) => (
										<div key={role} className="contents">
											<dt>{t(`item.roles.${role}`)}</dt>
											<dd>
												{entities[role].map((e, i) => (
													<span key={e.id}>
														{i > 0 && ", "}
														<Link
															href={localized(
																`/${ENTITY_PATH[e.kind]}/${e.code ?? e.slug}`,
																locale,
															)}
														>
															{e.name}
														</Link>
														{e.character && ` (${e.character})`}
													</span>
												))}
											</dd>
										</div>
									))}
								</dl>
							</section>
						</div>

						{trailers.length > 0 && (
							<section aria-labelledby="trailers-heading" className="grid gap-3">
								<h2 id="trailers-heading" className="text-xl font-bold">
									{t("item.trailers")}
								</h2>
								<Trailers trailers={trailers.slice(0, 3)} />
							</section>
						)}

						{loggedIn && (
							<StatusPicker
								itemId={item.id}
								type={item.type}
								status={me?.status ?? null}
							/>
						)}

						{timeToBeat && (
							<TimeToBeat
								itemId={item.id}
								medians={timeToBeat}
								mine={me?.timeToBeat}
								canReport={me?.status === "completed"}
							/>
						)}

						{editions && (
							<Editions slug={item.slug} editions={editions} total={editionCount} />
						)}

						{loggedIn && myLists && <AddToList itemId={item.id} lists={myLists} />}
					</div>
				</div>

				<p className="text-sm text-muted">{t("item.attribution")}</p>
			</article>
		</>
	);
}
