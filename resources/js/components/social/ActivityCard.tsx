import { Link, usePage } from "@inertiajs/react";

import { localized } from "../../lib/route";
import { useT } from "../../lib/useT";
import type { SharedProps } from "../../types";
import type { Cover } from "../CoverThumb";
import type { ItemRef } from "../library/ItemLink";
import ShelfCard from "../library/ShelfCard";

/** One entry of the `events` prop of Feed/Index (app/social/activity.go Entry). */
export interface ActivityEntry {
	id: string;
	kind: "rating_saved" | "status_changed" | "list_created" | "list_item_added";
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

	return (
		<article className="flex flex-col gap-2 py-4">
			<p className="text-sm">
				<Link href={localized(profile, locale)} className="font-semibold underline">
					{entry.actor.nickname}
				</Link>{" "}
				<span>{t(`social.feed.${entry.kind}`, params)}</span>{" "}
				{entry.list && (
					<Link
						href={localized(`${profile}/lists/${entry.list.id}`, locale)}
						className="underline"
					>
						{entry.list.title}
					</Link>
				)}
				<time dateTime={entry.at} className="ml-2 text-muted">
					{new Date(entry.at).toLocaleDateString(locale)}
				</time>
			</p>
			{entry.item && (
				<ShelfCard
					item={{ ...entry.item, year: null }}
					stars={entry.item.stars ?? undefined}
				/>
			)}
		</article>
	);
}
