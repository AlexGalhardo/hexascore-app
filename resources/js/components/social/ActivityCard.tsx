import { Link, usePage } from "@inertiajs/react";

import { localized } from "../../lib/route";
import { useT } from "../../lib/useT";
import type { SharedProps } from "../../types";
import type { Cover } from "../CoverThumb";
import type { ItemRef } from "../library/ItemLink";
import ShelfCard from "../library/ShelfCard";
import { formatScore } from "../ui/ScoreBadge";

/** One entry of the `events` prop of Feed/Index (app/social/activity.go Entry). */
export interface ActivityEntry {
	id: string;
	kind:
		| "rating_saved"
		| "status_changed"
		| "list_created"
		| "list_item_added"
		| "review_published";
	actor: { nickname: string; initial: string };
	item?: {
		id: string;
		type: ItemRef["type"];
		slug: string;
		title: string;
		cover: Cover | null;
		stars?: number;
	};
	list?: { id: string; title: string };
	payload: { status?: string };
	at: string;
}

export default function ActivityCard({ entry }: { entry: ActivityEntry }) {
	const t = useT();
	const { locale } = usePage<SharedProps>().props;
	const profile = `/u/${entry.actor.nickname.toLowerCase()}`;
	const params: Record<string, string> = { nickname: entry.actor.nickname };
	if (entry.kind === "status_changed" && entry.item && entry.payload.status) {
		params.status = t(`social.status.${entry.item.type}.${entry.payload.status}`);
	}

	const link =
		"font-semibold hover:text-primary focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fg";

	return (
		<article className="flex gap-3 py-5">
			<span
				aria-hidden="true"
				className="inline-flex size-9 shrink-0 items-center justify-center rounded-full bg-fg font-bold text-bg"
			>
				{entry.actor.initial}
			</span>
			<div className="min-w-0 flex-1">
				<p className="pt-1.5 text-[15px] leading-6">
					<Link href={localized(profile, locale)} className={link}>
						{entry.actor.nickname}
					</Link>{" "}
					<span>{t(`social.feed.${entry.kind}`, params)}</span>{" "}
					{entry.list && (
						<Link
							href={localized(`${profile}/lists/${entry.list.id}`, locale)}
							className={`${link} text-primary underline underline-offset-2`}
						>
							{entry.list.title}
						</Link>
					)}
					<time dateTime={entry.at} className="ml-2 text-sm text-muted tabular-nums">
						{new Date(entry.at).toLocaleDateString(locale)}
					</time>
				</p>
				{entry.item && (
					<div className="mt-2 max-w-md">
						{/* ShelfCard's `stars` reads "Your score": here the score is the actor's. */}
						<ShelfCard item={{ ...entry.item, year: null }}>
							{entry.item.stars !== undefined && (
								<p className="mt-2 text-sm font-semibold tabular-nums">
									{t("social.feed.score", {
										score: formatScore(entry.item.stars, locale),
									})}
								</p>
							)}
						</ShelfCard>
					</div>
				)}
			</div>
		</article>
	);
}
