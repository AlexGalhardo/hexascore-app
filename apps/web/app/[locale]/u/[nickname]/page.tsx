import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { getFormatter, getTranslations, setRequestLocale } from "next-intl/server";
import { cache } from "react";
import { z } from "zod";
import { CategoryLabel } from "@/components/catalog/category-label";
import { CoverThumb } from "@/components/catalog/cover-thumb";
import { ListForm } from "@/components/lists/list-form";
import { ScoreBadge } from "@/components/score/score-badge";
import { FollowButton } from "@/components/social/follow-button";
import { ReportDialog } from "@/components/social/report-dialog";
import { buttonClass } from "@/components/ui/form";
import { asLocale } from "@/i18n/locale";
import { Link } from "@/i18n/navigation";
import type { Locale } from "@/i18n/routing";
import { db } from "@/lib/db/client";
import { ITEM_TYPES, type ItemType } from "@/lib/db/schema/catalog";
import { itemHref } from "@/lib/sections";
import { getSessionUser } from "@/lib/session";
import { ITEM_STATUSES, LIST_TYPES, type ListType } from "@/lib/social/constants";
import { getUserLists } from "@/lib/social/lists";
import {
	getProfile,
	getProfileRatings,
	getProfileStats,
	getStatusShelves,
	isFollowing,
	PROFILE_SORTS,
	type ProfileRating,
} from "@/lib/social/repository";

type Props = {
	params: Promise<{ locale: string; nickname: string }>;
	searchParams: Promise<Record<string, string | string[] | undefined>>;
};

const querySchema = z.object({
	tab: z.enum(["ratings", "status", "lists"]).catch("ratings"),
	sort: z.enum(PROFILE_SORTS).catch("recent"),
	type: z.enum(ITEM_TYPES).optional().catch(undefined),
	page: z.coerce.number().int().min(1).max(500).catch(1),
});

/** Rows per category on the ratings tab overview (no `type` filter). */
const RATINGS_PREVIEW = 5;

const loadProfile = cache((nickname: string) =>
	nickname.length > 64 ? Promise.resolve(null) : getProfile(db, decodeURIComponent(nickname)),
);

export async function generateMetadata({ params }: Props): Promise<Metadata> {
	const { nickname } = await params;
	const profile = await loadProfile(nickname);
	if (!profile) return {};
	return { title: profile.nickname, robots: profile.isPrivate ? { index: false } : undefined };
}

export default async function PublicProfilePage({ params, searchParams }: Props) {
	const { locale, nickname } = await params;
	setRequestLocale(locale);
	const profile = await loadProfile(nickname);
	if (!profile) notFound();
	const [me, t, tcat, ts, format] = await Promise.all([
		getSessionUser(),
		getTranslations("social"),
		getTranslations("catalog"),
		getTranslations("score"),
		getFormatter(),
	]);
	const isOwner = me?.id === profile.id;
	const avatar = (
		<span
			aria-hidden="true"
			className="grid size-16 shrink-0 place-items-center rounded-full bg-fg font-serif text-3xl font-bold text-bg"
		>
			{profile.nickname.charAt(0).toUpperCase()}
		</span>
	);

	// Private (or pending deletion): same 200 page for everyone but the owner, with nothing beyond the nickname.
	if (profile.isPrivate && !isOwner) {
		return (
			<div className="mx-auto flex max-w-2xl items-center gap-4">
				{avatar}
				<div>
					<h1 className="font-serif text-3xl font-bold">{profile.nickname}</h1>
					<p className="text-muted">{t("profile.private")}</p>
				</div>
			</div>
		);
	}

	const query = querySchema.parse(await searchParams);
	const [stats, following] = await Promise.all([
		getProfileStats(db, profile.id),
		me && !isOwner ? isFollowing(db, me.id, profile.id) : Promise.resolve(false),
	]);
	const base = `/u/${encodeURIComponent(profile.nickname)}`;
	const tabHref = (tab: string, extra: Record<string, string> = {}) => ({
		pathname: base,
		query: { ...(tab === "ratings" ? {} : { tab }), ...extra },
	});
	const categoryHeading = (type: ItemType, id: string, count?: number) => (
		<h2 id={id} className="font-serif text-lg font-bold">
			<CategoryLabel type={type} size="lg">
				{tcat(`title.${type}`)}
			</CategoryLabel>
			{count === undefined ? null : <span className="ml-2 font-mono text-sm text-muted">{count}</span>}
		</h2>
	);

	return (
		<div className="grid grid-cols-1 gap-8">
			<header className="grid gap-4 sm:grid-cols-[auto_1fr_auto] sm:items-start">
				{avatar}
				<div className="grid min-w-0 gap-1">
					<h1 className="truncate font-serif text-3xl font-bold">{profile.nickname}</h1>
					<p className="text-sm text-muted">
						{t("profile.memberSince", {
							date: format.dateTime(new Date(profile.memberSince), { dateStyle: "long" }),
						})}
						{profile.isPrivate ? ` · ${t("profile.privateOwner")}` : null}
					</p>
					<p className="flex flex-wrap gap-x-3 gap-y-1 text-sm">
						{ITEM_TYPES.map((type) => (
							<CategoryLabel key={type} type={type}>
								{t(`profile.counts.${type}`, { count: stats.byType[type] })}
							</CategoryLabel>
						))}
					</p>
					<p className="text-sm">
						{t("profile.followers", { count: stats.followers })}
						{" · "}
						{t("profile.following", { count: stats.following })}
					</p>
				</div>
				<div className="flex flex-wrap items-center gap-3 sm:justify-end">
					<div className="flex items-center gap-2">
						<ScoreBadge
							score={stats.averageOverall}
							size="md"
							label={
								stats.averageOverall === null
									? ts("none")
									: ts("outOf", { score: stats.averageOverall })
							}
						/>
						<span className="text-sm text-muted">{t("profile.averageGiven")}</span>
					</div>
					{me && !isOwner ? <FollowButton nickname={profile.nickname} initial={following} /> : null}
					{!me ? (
						<Link href={{ pathname: "/login", query: { next: base } }} className={buttonClass("primary")}>
							{t("profile.follow")}
						</Link>
					) : null}
					{me && !isOwner ? <ReportDialog target={{ type: "user", nickname: profile.nickname }} /> : null}
				</div>
			</header>

			<nav aria-label={t("profile.sections")} className="flex gap-1 border-b border-border">
				{(["ratings", "status", "lists"] as const).map((tab) => (
					<Link
						key={tab}
						href={tabHref(tab)}
						aria-current={query.tab === tab ? "page" : undefined}
						className="-mb-px grid min-h-11 place-items-center border-b-2 border-transparent px-3 text-sm aria-[current=page]:border-fg aria-[current=page]:font-semibold"
					>
						{t(`profile.tabs.${tab}`)}
					</Link>
				))}
			</nav>

			{query.tab === "ratings" ? (
				<RatingsTab
					userId={profile.id}
					locale={asLocale(locale)}
					sort={query.sort}
					page={query.page}
					type={query.type}
					counts={stats.byType}
				/>
			) : null}
			{query.tab === "status" ? <StatusTab userId={profile.id} locale={asLocale(locale)} /> : null}
			{query.tab === "lists" ? (
				<ListsTab userId={profile.id} nickname={profile.nickname} isOwner={isOwner} />
			) : null}
		</div>
	);

	function RatingRow({ rating }: { rating: ProfileRating }) {
		return (
			<li className="flex items-center gap-3 border-t border-border py-2">
				<CoverThumb src={rating.item.coverUrl} thumb={rating.item.coverThumbUrl} />
				<div className="min-w-0 flex-1">
					<Link href={itemHref(rating.item.type, rating.item.slug)} className="block truncate font-medium">
						{rating.item.title}
					</Link>
					<p className="text-xs text-muted">
						{format.dateTime(new Date(rating.ratedAt), { dateStyle: "medium" })}
					</p>
				</div>
				<ScoreBadge score={rating.overall} size="sm" label={ts("outOf", { score: rating.overall })} />
			</li>
		);
	}

	/**
	 * Listing pattern: without `type`, one section per rated category with its first RATINGS_PREVIEW rows (one small
	 * query per category) and a "see all" link; with `type`, that category alone with the usual offset pagination.
	 */
	async function RatingsTab(props: {
		userId: string;
		locale: Locale;
		sort: (typeof PROFILE_SORTS)[number];
		page: number;
		type: ItemType | undefined;
		counts: Record<ItemType, number>;
	}) {
		const sortQuery: Record<string, string> = props.sort === "recent" ? {} : { sort: props.sort };
		const typeQuery: Record<string, string> = props.type ? { type: props.type } : {};
		const link = (extra: Record<string, string>) => tabHref("ratings", extra);
		const shown = props.type ? [props.type] : ITEM_TYPES.filter((type) => props.counts[type] > 0);
		const groups = await Promise.all(
			shown.map(async (type) => ({
				type,
				...(await getProfileRatings(db, props.userId, {
					sort: props.sort,
					locale: props.locale,
					type,
					page: props.type ? props.page : 1,
					limit: props.type ? undefined : RATINGS_PREVIEW,
				})),
			})),
		);
		const visible = groups.filter((g) => g.data.length);
		return (
			<section className="grid gap-4" aria-label={t("profile.tabs.ratings")}>
				<nav aria-label={t("profile.byCategory")} className="flex flex-wrap gap-2 text-sm">
					{[undefined, ...ITEM_TYPES].map((type) => (
						<Link
							key={type ?? "all"}
							href={link({ ...sortQuery, ...(type ? { type } : {}) })}
							aria-current={props.type === type ? "page" : undefined}
							className="inline-flex min-h-11 items-center rounded-md border border-border px-3 aria-[current=page]:bg-surface aria-[current=page]:font-semibold"
						>
							{type ? (
								<CategoryLabel type={type}>{tcat(`title.${type}`)}</CategoryLabel>
							) : (
								t("profile.allTypes")
							)}
						</Link>
					))}
				</nav>
				<div className="flex flex-wrap items-center gap-2 text-sm">
					<span className="text-muted">{t("profile.sortBy")}</span>
					{PROFILE_SORTS.map((s) => (
						<Link
							key={s}
							href={link({ ...typeQuery, ...(s === "recent" ? {} : { sort: s }) })}
							aria-current={props.sort === s ? "true" : undefined}
							className="grid min-h-11 place-items-center rounded-md px-2 aria-[current=true]:bg-surface aria-[current=true]:font-semibold"
						>
							{t(`profile.sorts.${s}`)}
						</Link>
					))}
				</div>
				{visible.length ? (
					visible.map((g) => (
						<section key={g.type} aria-labelledby={`ratings-${g.type}`} className="grid gap-1">
							{categoryHeading(g.type, `ratings-${g.type}`, props.counts[g.type])}
							<ol className="grid">
								{g.data.map((rating) => (
									<RatingRow key={rating.item.id} rating={rating} />
								))}
							</ol>
							{!props.type && g.hasMore ? (
								<Link
									href={link({ ...sortQuery, type: g.type })}
									className="inline-flex min-h-11 items-center justify-self-start text-sm font-medium underline"
								>
									{t("profile.seeAll", { count: props.counts[g.type] })}
								</Link>
							) : null}
						</section>
					))
				) : (
					<p className="text-muted">{t("profile.noRatings")}</p>
				)}
				{props.type ? (
					<div className="flex gap-2">
						{props.page > 1 ? (
							<Link
								href={link({ ...typeQuery, ...sortQuery, page: String(props.page - 1) })}
								className={buttonClass("secondary")}
							>
								{t("profile.previous")}
							</Link>
						) : null}
						{groups[0]?.hasMore ? (
							<Link
								href={link({ ...typeQuery, ...sortQuery, page: String(props.page + 1) })}
								className={buttonClass("secondary")}
							>
								{t("profile.next")}
							</Link>
						) : null}
					</div>
				) : null}
			</section>
		);
	}

	/** Listing pattern: one section per category holding its non-empty shelves, with that category's status labels. */
	async function StatusTab({ userId, locale }: { userId: string; locale: Locale }) {
		const shelves = await getStatusShelves(db, userId, locale);
		const groups = ITEM_TYPES.map((type) => ({
			type,
			shelves: ITEM_STATUSES.map((status) => ({
				status,
				items: shelves[status].filter((item) => item.type === type),
			})).filter((shelf) => shelf.items.length),
		})).filter((g) => g.shelves.length);
		if (!groups.length) return <p className="text-muted">{t("profile.emptyShelf")}</p>;
		return (
			<div className="grid gap-8">
				{groups.map((g) => (
					<section key={g.type} aria-labelledby={`status-${g.type}`} className="grid gap-3">
						{categoryHeading(g.type, `status-${g.type}`)}
						<div className="grid gap-6 md:grid-cols-2">
							{g.shelves.map(({ status, items }) => (
								<section
									key={status}
									aria-labelledby={`shelf-${g.type}-${status}`}
									className="grid content-start gap-1"
								>
									<h3 id={`shelf-${g.type}-${status}`} className="text-sm font-semibold">
										{t(`status.${g.type}.${status}`)}{" "}
										<span className="font-mono text-muted">{items.length}</span>
									</h3>
									<ul className="grid">
										{items.map((item) => (
											<li
												key={item.id}
												className="flex items-center gap-3 border-t border-border py-1.5"
											>
												<CoverThumb src={item.coverUrl} thumb={item.coverThumbUrl} />
												<Link
													href={itemHref(item.type, item.slug)}
													className="min-w-0 truncate"
												>
													{item.title}
												</Link>
											</li>
										))}
									</ul>
								</section>
							))}
						</div>
					</section>
				))}
			</div>
		);
	}

	async function ListsTab({ userId, nickname, isOwner }: { userId: string; nickname: string; isOwner: boolean }) {
		const [lists, tl] = await Promise.all([getUserLists(db, userId, isOwner), getTranslations("lists")]);
		const groups = LIST_TYPES.map((type) => ({ type, lists: lists.filter((l) => l.type === type) })).filter(
			(g) => g.lists.length,
		);
		return (
			<div className="grid gap-6">
				{groups.length ? (
					groups.map((g) => (
						<section key={g.type} aria-labelledby={`lists-${g.type}`} className="grid gap-1">
							{categoryHeading(g.type, `lists-${g.type}`)}
							<ul className="grid">
								{g.lists.map((l) => (
									<li key={l.id} className="grid gap-1 border-t border-border py-3">
										<Link
											href={`/u/${encodeURIComponent(nickname)}/lists/${l.id}`}
											className="font-medium"
										>
											{l.title}
										</Link>
										<p className="text-xs text-muted">
											{[
												tl("itemCount", { count: l.itemCount }),
												isOwner ? (l.isPublic ? tl("public") : tl("private")) : null,
												l.hidden ? tl("hidden") : null,
											]
												.filter(Boolean)
												.join(" · ")}
										</p>
										{l.description ? <p className="text-sm">{l.description}</p> : null}
									</li>
								))}
							</ul>
						</section>
					))
				) : (
					<p className="text-muted">{isOwner ? tl("noLists") : tl("noPublicLists")}</p>
				)}
				{isOwner ? (
					<section aria-labelledby="new-list" className="grid gap-3 border-t border-border pt-4">
						<h2 id="new-list" className="font-serif text-lg font-bold">
							{tl("new")}
						</h2>
						<ListForm
							counts={
								Object.fromEntries(
									LIST_TYPES.map((type) => [type, lists.filter((l) => l.type === type).length]),
								) as Record<ListType, number>
							}
						/>
					</section>
				) : null}
			</div>
		);
	}
}
