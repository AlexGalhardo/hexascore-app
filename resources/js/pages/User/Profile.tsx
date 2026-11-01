import { InfiniteScroll, Link, usePage } from "@inertiajs/react";

import type { ItemRef } from "../../components/library/ItemLink";
import ShelfCard from "../../components/library/ShelfCard";
import ReportDialog from "../../components/ReportDialog";
import Seo from "../../components/Seo";
import FollowBar, { type SocialProps } from "../../components/social/FollowBar";
import { localized } from "../../lib/route";
import { useT } from "../../lib/useT";
import type { SharedProps } from "../../types";

const TABS = ["ratings", "status", "lists"] as const;
type Tab = (typeof TABS)[number];

interface Rating extends ItemRef {
	stars: number;
	ratedAt: string;
}

interface ProfileList {
	id: string;
	type: ItemRef["type"];
	title: string;
	itemCount: number;
}

interface Props {
	profile: { nickname: string; initial: string; isPrivate: boolean; isOwner: boolean };
	stats?: {
		ratingsByType: Record<string, number>;
		statusByType: Record<string, Record<string, number>>;
	};
	ratings?: { data: Rating[] };
	lists?: ProfileList[];
	indexable: boolean;
	social?: SocialProps;
	contributions?: {
		submitted: number;
		approved: number;
		rejected: number;
		approvalRate: number;
		tier?: string;
		accuracy?: number;
	};
}

export default function Profile({
	profile,
	stats,
	ratings,
	lists,
	indexable,
	social,
	contributions,
}: Props) {
	const t = useT();
	const page = usePage<SharedProps>();
	const viewer = page.props.auth?.user;
	const base = localized(`/u/${profile.nickname.toLowerCase()}`, page.props.locale);
	const requested = new URLSearchParams(page.url.split("?")[1] ?? "").get("tab");
	const tab: Tab = TABS.find((x) => x === requested) ?? "ratings";

	return (
		<>
			<Seo title={profile.nickname} noindex={!indexable} />
			<header className="flex flex-wrap items-center gap-4">
				<span
					aria-hidden="true"
					className="flex size-16 items-center justify-center rounded-full bg-accent-game text-2xl font-semibold text-white"
				>
					{profile.initial}
				</span>
				<div className="min-w-0 flex-1">
					<h1 className="text-2xl font-semibold">{profile.nickname}</h1>
					{contributions && (
						<section
							aria-label={t("social.profile.contributions.title")}
							className="text-sm text-muted"
						>
							<p>
								{t("social.profile.contributions.summary", {
									submitted: contributions.submitted,
									approved: contributions.approved,
									rejected: contributions.rejected,
								})}
								{" · "}
								{t("social.profile.contributions.rate", {
									rate: new Intl.NumberFormat(page.props.locale, {
										style: "percent",
									}).format(contributions.approvalRate),
								})}
							</p>
							{contributions.tier && (
								<p>
									{t("social.profile.contributions.tier", {
										tier: contributions.tier,
										accuracy: new Intl.NumberFormat(page.props.locale, {
											style: "percent",
										}).format(contributions.accuracy ?? 0),
									})}
								</p>
							)}
						</section>
					)}
					{stats && (
						<p className="text-sm text-muted">
							{t("social.profile.summary", {
								ratings: Object.values(stats.ratingsByType).reduce(
									(a, b) => a + b,
									0,
								),
								lists: lists?.length ?? 0,
							})}
						</p>
					)}
				</div>
				{viewer && !profile.isOwner && (
					<ReportDialog targetType="user" targetId={profile.nickname} />
				)}
			</header>
			{social && (
				<FollowBar
					nickname={profile.nickname}
					social={social}
					canFollow={Boolean(viewer) && !profile.isOwner}
				/>
			)}
			{profile.isPrivate && !profile.isOwner && (
				<p className="mt-4">{t("social.profile.private")}</p>
			)}
			{profile.isPrivate && profile.isOwner && (
				<p className="mt-4 text-muted">{t("social.profile.privateOwner")}</p>
			)}
			{stats && (
				<>
					<nav
						aria-label={t("social.profile.sections")}
						className="mt-4 flex gap-4 border-b border-border"
					>
						{TABS.map((x) => (
							<Link
								key={x}
								href={`${base}?tab=${x}`}
								preserveState
								aria-current={x === tab ? "page" : undefined}
								className={`pb-2 ${x === tab ? "border-b-2 border-current font-semibold" : "text-muted"}`}
							>
								{t(`social.profile.tabs.${x}`)}
							</Link>
						))}
					</nav>
					{tab === "ratings" && (
						<section className="mt-4">
							{(ratings?.data.length ?? 0) === 0 ? (
								<p>{t("social.profile.noRatings")}</p>
							) : (
								<InfiniteScroll
									data="ratings"
									as="ul"
									className="divide-y divide-border"
								>
									{ratings?.data.map((r) => (
										<li key={r.id} className="py-2">
											<ShelfCard
												item={r}
												stars={r.stars}
												statusLabel={t(`itemTypes.${r.type}`)}
											/>
										</li>
									))}
								</InfiniteScroll>
							)}
						</section>
					)}
					{tab === "status" && (
						<section className="mt-4 space-y-4">
							{Object.entries(stats.statusByType).map(([type, byStatus]) => (
								<div key={type}>
									<h2 className="font-semibold">{t(`itemTypes.${type}`)}</h2>
									<ul className="mt-1 text-sm">
										{Object.entries(byStatus).map(([status, n]) => (
											<li key={status}>
												{t(`social.status.${type}.${status}`)}: {n}
											</li>
										))}
									</ul>
								</div>
							))}
							{Object.keys(stats.statusByType).length === 0 && (
								<p>{t("social.profile.emptyShelf")}</p>
							)}
						</section>
					)}
					{tab === "lists" && (
						<section className="mt-4">
							{(lists?.length ?? 0) === 0 ? (
								<p>{t("social.profile.emptyShelf")}</p>
							) : (
								<ul className="divide-y divide-border">
									{lists?.map((l) => (
										<li key={l.id} className="py-2">
											<Link
												href={`${base}/lists/${l.id}`}
												className="font-medium underline"
											>
												{l.title}
											</Link>{" "}
											<span className="text-sm text-muted">
												{t("userLists.itemCount", { count: l.itemCount })}
											</span>
										</li>
									))}
								</ul>
							)}
						</section>
					)}
				</>
			)}
		</>
	);
}
