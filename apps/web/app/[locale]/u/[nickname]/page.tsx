import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { getFormatter, getTranslations, setRequestLocale } from "next-intl/server";
import { cache } from "react";
import { z } from "zod";
import { ListForm } from "@/components/lists/list-form";
import { ScoreBadge } from "@/components/score/score-badge";
import { FollowButton } from "@/components/social/follow-button";
import { ReportDialog } from "@/components/social/report-dialog";
import { buttonClass } from "@/components/ui/form";
import { Link } from "@/i18n/navigation";
import type { Locale } from "@/i18n/routing";
import { db } from "@/lib/db/client";
import { ITEM_TYPES } from "@/lib/db/schema/catalog";
import { itemHref } from "@/lib/sections";
import { getSessionUser } from "@/lib/session";
import { ITEM_STATUSES } from "@/lib/social/constants";
import { getUserLists } from "@/lib/social/lists";
import {
	getProfile,
	getProfileRatings,
	getProfileStats,
	getStatusShelves,
	isFollowing,
	PROFILE_SORTS,
} from "@/lib/social/repository";

type Props = {
	params: Promise<{ locale: string; nickname: string }>;
	searchParams: Promise<Record<string, string | string[] | undefined>>;
};

const querySchema = z.object({
	tab: z.enum(["ratings", "status", "lists"]).catch("ratings"),
	sort: z.enum(PROFILE_SORTS).catch("recent"),
	page: z.coerce.number().int().min(1).max(500).catch(1),
});

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
	const [me, t, tt, ts, format] = await Promise.all([
		getSessionUser(),
		getTranslations("social"),
		getTranslations("itemTypes"),
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

	return (
		<div className="grid gap-8">
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
					<p className="text-sm">
						{ITEM_TYPES.map((type) => t(`profile.counts.${type}`, { count: stats.byType[type] })).join(
							" · ",
						)}
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
					locale={locale as Locale}
					sort={query.sort}
					page={query.page}
					tabHref={tabHref}
				/>
			) : null}
			{query.tab === "status" ? <StatusTab userId={profile.id} locale={locale as Locale} /> : null}
			{query.tab === "lists" ? (
				<ListsTab userId={profile.id} nickname={profile.nickname} isOwner={isOwner} />
			) : null}
		</div>
	);

	async function RatingsTab(props: {
		userId: string;
		locale: Locale;
		sort: (typeof PROFILE_SORTS)[number];
		page: number;
		tabHref: (tab: string, extra?: Record<string, string>) => { pathname: string; query: Record<string, string> };
	}) {
		const { data, hasMore } = await getProfileRatings(db, props.userId, props);
		const link = (extra: Record<string, string>) => props.tabHref("ratings", extra);
		return (
			<section className="grid gap-4" aria-label={t("profile.tabs.ratings")}>
				<div className="flex flex-wrap items-center gap-2 text-sm">
					<span className="text-muted">{t("profile.sortBy")}</span>
					{PROFILE_SORTS.map((s) => (
						<Link
							key={s}
							href={link(s === "recent" ? {} : { sort: s })}
							aria-current={props.sort === s ? "true" : undefined}
							className="grid min-h-11 place-items-center rounded-md px-2 aria-[current=true]:bg-surface aria-[current=true]:font-semibold"
						>
							{t(`profile.sorts.${s}`)}
						</Link>
					))}
				</div>
				{data.length ? (
					<ol className="grid">
						{data.map((r) => (
							<li
								key={r.item.id}
								className="grid grid-cols-[auto_1fr] items-center gap-3 border-t border-border py-2"
							>
								<ScoreBadge score={r.overall} size="sm" label={ts("outOf", { score: r.overall })} />
								<div className="min-w-0">
									<Link
										href={itemHref(r.item.type, r.item.slug)}
										className="block truncate font-medium hover:underline"
									>
										{r.item.title}
									</Link>
									<p className="text-xs text-muted">
										{tt(r.item.type)} ·{" "}
										{format.dateTime(new Date(r.ratedAt), { dateStyle: "medium" })}
									</p>
								</div>
							</li>
						))}
					</ol>
				) : (
					<p className="text-muted">{t("profile.noRatings")}</p>
				)}
				<div className="flex gap-2">
					{props.page > 1 ? (
						<Link
							href={link({
								...(props.sort === "recent" ? {} : { sort: props.sort }),
								page: String(props.page - 1),
							})}
							className={buttonClass("secondary")}
						>
							{t("profile.previous")}
						</Link>
					) : null}
					{hasMore ? (
						<Link
							href={link({
								...(props.sort === "recent" ? {} : { sort: props.sort }),
								page: String(props.page + 1),
							})}
							className={buttonClass("secondary")}
						>
							{t("profile.next")}
						</Link>
					) : null}
				</div>
			</section>
		);
	}

	async function StatusTab({ userId, locale }: { userId: string; locale: Locale }) {
		const shelves = await getStatusShelves(db, userId, locale);
		return (
			<div className="grid gap-6 md:grid-cols-2">
				{ITEM_STATUSES.map((status) => (
					<section key={status} aria-labelledby={`shelf-${status}`} className="grid content-start gap-2">
						<h2 id={`shelf-${status}`} className="font-serif text-lg font-bold">
							{t(`status.shelves.${status}`)}{" "}
							<span className="font-mono text-sm text-muted">{shelves[status].length}</span>
						</h2>
						{shelves[status].length ? (
							<ul className="grid">
								{shelves[status].map((item) => (
									<li
										key={item.id}
										className="flex min-h-11 items-center gap-2 border-t border-border"
									>
										<Link
											href={itemHref(item.type, item.slug)}
											className="truncate hover:underline"
										>
											{item.title}
										</Link>
										<span className="ml-auto shrink-0 text-xs text-muted">{tt(item.type)}</span>
									</li>
								))}
							</ul>
						) : (
							<p className="text-sm text-muted">{t("profile.emptyShelf")}</p>
						)}
					</section>
				))}
			</div>
		);
	}

	async function ListsTab({ userId, nickname, isOwner }: { userId: string; nickname: string; isOwner: boolean }) {
		const [lists, tl] = await Promise.all([getUserLists(db, userId, isOwner), getTranslations("lists")]);
		return (
			<div className="grid gap-6">
				{lists.length ? (
					<ul className="grid">
						{lists.map((l) => (
							<li key={l.id} className="grid gap-1 border-t border-border py-3">
								<Link
									href={`/u/${encodeURIComponent(nickname)}/lists/${l.id}`}
									className="font-medium hover:underline"
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
				) : (
					<p className="text-muted">{isOwner ? tl("noLists") : tl("noPublicLists")}</p>
				)}
				{isOwner ? (
					<section aria-labelledby="new-list" className="grid gap-3 border-t border-border pt-4">
						<h2 id="new-list" className="font-serif text-lg font-bold">
							{tl("new")}
						</h2>
						<ListForm />
					</section>
				) : null}
			</div>
		);
	}
}
