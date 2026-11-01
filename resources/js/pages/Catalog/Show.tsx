import { Link, usePage } from "@inertiajs/react";
import { lazy, type ReactNode, Suspense, useState } from "react";
import CategoryIcon from "../../components/CategoryIcon";
import CoverThumb, { type Cover } from "../../components/CoverThumb";
import {
	ChartIcon,
	HistoryIcon,
	ListIcon,
	PencilIcon,
	StarIcon,
} from "../../components/item/ActionIcons";
import AddToList, { type ListChoice } from "../../components/item/AddToList";
import CategoryAverages from "../../components/item/CategoryAverages";
import Editions from "../../components/item/Editions";
import FriendsScore, { type FriendsScoreData } from "../../components/item/FriendsScore";
import MoreLikeThis, { type CrossItem, type SimilarItem } from "../../components/item/MoreLikeThis";
import RatingPanel from "../../components/item/RatingPanel";
import Reviews, { type ReviewEntry } from "../../components/item/Reviews";
import ScoreBlock from "../../components/item/ScoreBlock";
import StatusPicker from "../../components/item/StatusPicker";
import TimeToBeat, { type Medians } from "../../components/item/TimeToBeat";
import Trailers from "../../components/item/Trailers";
import {
	CATEGORY_KEYS,
	type EditionRow,
	type ItemType,
	type Me,
	type RatingSummary,
} from "../../components/item/types";
import Seo, { truncate } from "../../components/Seo";
import ShareButton from "../../components/social/ShareButton";
import { btnClass } from "../../components/ui/Btn";
import { localized } from "../../lib/route";
import { useT } from "../../lib/useT";
import type { SharedProps } from "../../types";

const RatingsDialog = lazy(() => import("../../components/item/RatingsDialog"));

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
	/** Deferred (group `reviews`) while the feature flag `social` is on (SDD 14 §12.10). */
	reviews?: { data: ReviewEntry[]; mine: ReviewEntry | null };
	/** Deferred (group `personal`) for signed-in viewers while the flag `social` is on. */
	friendsScore?: FriendsScoreData | null;
	/** Share-card path while the feature flag `social` is on and the page is public (SDD 14 §11 item 12). */
	ogImage?: string;
	similar?: SimilarItem[];
	crossMedia?: CrossItem[];
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

// Chips under the title: genres and platforms, linked to their entity pages.
const CHIP_ROLES = ["genre", "platform"];
// The first of these credits completes the meta line under the title.
const CREDIT_ROLES = ["developer", "author", "director", "creator", "network"];

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
	reviews,
	friendsScore,
	ogImage,
	similar = [],
	crossMedia = [],
}: Props) {
	const t = useT();
	const { locale, auth } = usePage<SharedProps>().props;
	const l = (path: string) => localized(path, locale);
	const title = item.title ?? item.slug;
	const loggedIn = Boolean(auth?.user);
	const roles = ROLE_ORDER.filter((r) => entities[r]?.length);
	const keys = CATEGORY_KEYS[item.type];
	const categoryLabels = keys.map((k) => t(`categories.${item.type}.${k}`));
	const [scoresOpen, setScoresOpen] = useState(false);
	// Entity pages live under the item's own category (owner round 2026-11-01 item 19).
	const entityHref = (e: EntityLink) =>
		l(`/${SECTIONS[item.type]}/${ENTITY_PATH[e.kind]}/${e.code ?? e.slug}`);
	const credit = CREDIT_ROLES.map((r) => entities[r]?.[0]).find(Boolean);
	const chips = CHIP_ROLES.flatMap((r) => entities[r] ?? []);
	const release =
		item.releaseDate?.length === 10
			? new Intl.DateTimeFormat(locale, { dateStyle: "long", timeZone: "UTC" }).format(
					new Date(`${item.releaseDate}T00:00:00Z`),
				)
			: item.releaseDate;

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

	const action = btnClass("secondary", "w-full justify-start");
	const seeScores = (
		<button
			type="button"
			className={action}
			onClick={() => setScoresOpen(true)}
			disabled={scoresOpen}
		>
			<ChartIcon />
			{t("item.seeUserScores")}
		</button>
	);

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
				image={ogImage}
			/>
			<nav aria-label={t("item.breadcrumb")} className="mb-6 text-sm text-muted">
				<ol className="flex flex-wrap gap-2">
					<li>
						<Link
							href={l(`/${SECTIONS[item.type]}`)}
							className="inline-flex items-center gap-1.5 hover:text-fg hover:underline"
						>
							<CategoryIcon type={item.type} className="size-4" />
							{t(`nav.${SECTIONS[item.type]}`)}
						</Link>
					</li>
					<li aria-hidden="true">/</li>
					<li aria-current="page" className="text-fg">
						{title}
					</li>
				</ol>
			</nav>
			{/* Owner request (round 2026-11-01 item 7): no mouse text selection on the item page. Form fields stay
			    selectable; keyboard focus and screen readers are unaffected by user-select. */}
			<article
				data-testid="item-page"
				className="grid gap-8 select-none [&_input]:select-text [&_textarea]:select-text sm:grid-cols-[12rem_minmax(0,1fr)] lg:grid-cols-[12rem_minmax(0,1fr)_19rem] lg:gap-x-10 xl:grid-cols-[14rem_minmax(0,1fr)_22rem]"
			>
				<div data-testid="item-col-actions" className="grid content-start gap-4">
					<CoverThumb cover={cover} alt={title} size="lg" priority />
					<div className="grid grid-cols-1 gap-2 min-[400px]:grid-cols-2 sm:grid-cols-1">
						{scoresOpen ? (
							<Suspense fallback={seeScores}>
								<RatingsDialog
									itemId={item.id}
									categoryLabels={categoryLabels}
									initialOpen
									icon={<ChartIcon />}
									triggerClassName={action}
								/>
							</Suspense>
						) : (
							seeScores
						)}
						<a
							href="#your-rating"
							className={btnClass("primary", "w-full justify-start")}
						>
							<StarIcon />
							{loggedIn && me?.rating ? t("item.goToRating") : t("item.rateIt")}
						</a>
						{loggedIn && myLists ? (
							<AddToList
								itemId={item.id}
								lists={myLists}
								icon={<ListIcon />}
								triggerClassName={action}
							/>
						) : (
							<Link href={l("/login")} className={action}>
								<ListIcon />
								{t("lists.addToList")}
							</Link>
						)}
						<Link href={l(`/suggest/item/${item.id}`)} className={action}>
							<PencilIcon />
							{t("community.suggest.link")}
						</Link>
						<Link
							href={l(`/${SECTIONS[item.type]}/${item.slug}/changes`)}
							className={action}
						>
							<HistoryIcon />
							{t("community.recent.link")}
						</Link>
					</div>
				</div>

				<div data-testid="item-col-main" className="grid min-w-0 content-start gap-10">
					<header className="grid content-start gap-4">
						<h1 className="text-3xl leading-none sm:text-4xl md:text-5xl">{title}</h1>
						{ogImage && <ShareButton title={title} />}
						<p className="text-muted md:text-lg">
							{[
								t(`itemTypes.${item.type}`),
								item.releaseDate?.slice(0, 4),
								credit?.name,
							]
								.filter(Boolean)
								.join(" · ")}
						</p>
						{chips.length > 0 && (
							<ul data-testid="item-chips" className="flex flex-wrap gap-2">
								{chips.map((e) => (
									<li key={`${e.kind}-${e.id}`}>
										<Link
											href={entityHref(e)}
											className="inline-flex min-h-10 items-center rounded-xs border border-fg px-3.5 text-sm hover:bg-hover focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary"
										>
											{e.name}
										</Link>
									</li>
								))}
							</ul>
						)}
					</header>

					<Section id="synopsis" title={t("item.synopsis")}>
						{item.synopsis ? (
							<>
								<p className="text-lg/relaxed">{item.synopsis.text}</p>
								<p className="text-xs text-muted">
									{t("item.license", { license: item.synopsis.licence })}
									{item.synopsis.sourceUrl && (
										<>
											{" · "}
											<a
												href={item.synopsis.sourceUrl}
												rel="noopener noreferrer nofollow"
												className="text-primary hover:underline"
											>
												{t("item.source")}
											</a>
										</>
									)}
								</p>
							</>
						) : (
							<p className="text-muted">{t("item.noSynopsis")}</p>
						)}
					</Section>

					<Section id="averages" title={t("item.categoryScores")}>
						<CategoryAverages
							keys={keys}
							labels={categoryLabels}
							averages={ratingSummary.scoreByCategory}
						/>
					</Section>

					<div className="grid gap-12 md:grid-cols-2 md:gap-8">
						<Section id="details" title={t("item.details")}>
							<dl className="grid">
								{release && (
									<DetailRow label={t("item.fields.releaseDate")}>
										{release}
									</DetailRow>
								)}
								{item.runtimeMinutes !== undefined && (
									<DetailRow label={t("item.fields.runtime")}>
										{t("item.runtimeValue", { minutes: item.runtimeMinutes })}
									</DetailRow>
								)}
								{item.pageCount !== undefined && (
									<DetailRow label={t("item.fields.pages")}>
										{item.pageCount}
									</DetailRow>
								)}
								{roles.map((role) => (
									<DetailRow key={role} label={t(`item.roles.${role}`)}>
										{entities[role].map((e, i) => (
											<span key={e.id}>
												{i > 0 && ", "}
												<Link
													href={entityHref(e)}
													className="text-primary hover:underline"
												>
													{e.name}
												</Link>
												{e.character && ` (${e.character})`}
											</span>
										))}
									</DetailRow>
								))}
							</dl>
						</Section>

						{timeToBeat && (
							<TimeToBeat
								itemId={item.id}
								medians={timeToBeat}
								mine={me?.timeToBeat}
								canReport={me?.status === "completed"}
							/>
						)}
					</div>

					{trailers.length > 0 && (
						<Section
							id="trailers"
							title={t("item.trailers")}
							count={Math.min(trailers.length, 3)}
						>
							<Trailers trailers={trailers.slice(0, 3)} />
						</Section>
					)}

					{editions && (
						<Editions slug={item.slug} editions={editions} total={editionCount} />
					)}

					{reviews && (
						<Section
							id="reviews"
							title={t("reviews.title")}
							count={reviews.data.length}
						>
							<Reviews
								itemId={item.id}
								reviews={reviews}
								canWrite={loggedIn}
								rated={Boolean(me?.rating)}
							/>
						</Section>
					)}

					<MoreLikeThis type={item.type} similar={similar} crossMedia={crossMedia} />

					<p className="text-xs text-muted">{t("item.attribution")}</p>
				</div>

				<aside
					data-testid="item-col-rating"
					className="grid content-start gap-4 sm:col-span-2 lg:col-span-1 lg:self-start"
				>
					<ScoreBlock
						score={ratingSummary.score}
						ratingCount={ratingSummary.ratingCount}
					/>
					<FriendsScore data={friendsScore} />
					<RatingPanel
						key={item.id + (me?.rating ? "r" : "n")}
						itemId={item.id}
						type={item.type}
						mine={me?.rating}
						canWrite={loggedIn}
					>
						<StatusPicker
							itemId={item.id}
							type={item.type}
							status={me?.status ?? null}
						/>
						{myLists && <AddToList itemId={item.id} lists={myLists} full />}
					</RatingPanel>
				</aside>
			</article>
		</>
	);
}

// A block of the item page: 2 px ink rule on top, serif heading (design v1).
function Section({
	id,
	title,
	count,
	children,
}: {
	id: string;
	title: string;
	count?: number;
	children: ReactNode;
}) {
	return (
		<section
			aria-labelledby={`${id}-heading`}
			className="grid content-start gap-4 border-t-2 border-fg pt-4"
		>
			<h2 id={`${id}-heading`} className="text-2xl">
				{title}
				{count !== undefined && (
					<span className="ml-2 align-middle font-sans text-xs font-normal text-muted">
						{count}
					</span>
				)}
			</h2>
			{children}
		</section>
	);
}

function DetailRow({ label, children }: { label: string; children: ReactNode }) {
	return (
		<div className="grid grid-cols-[7rem_minmax(0,1fr)] gap-4 border-b border-border py-2.5 text-sm">
			<dt className="text-muted">{label}</dt>
			<dd>{children}</dd>
		</div>
	);
}
