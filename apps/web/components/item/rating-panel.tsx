import { getTranslations } from "next-intl/server";
import type { ReactNode } from "react";
import { Link } from "@/i18n/navigation";
import { db } from "@/lib/db/client";
import type { ItemType } from "@/lib/db/schema/catalog";
import { getUserRating } from "@/lib/ratings";
import { getSessionUser } from "@/lib/session";
import { getListsForItem } from "@/lib/social/lists";
import { getItemStatus } from "@/lib/social/repository";
import { getUserTimeToBeat } from "@/lib/time-to-beat";
import { buttonClass, Notice } from "../ui/form";
import { AddToList } from "./add-to-list";
import { ItemStatusSelect } from "./item-status-select";
import { RatingForm } from "./rating-form";
import { TimeToBeatForm } from "./time-to-beat-form";

type Props = { itemId: string; type: ItemType; labels: string[]; path: string };

/** The viewer's rating, status and lists for an item. Per-user, so read uncached (the item page is dynamic). */
export async function RatingPanel({ itemId, type, labels, path }: Props) {
	const [t, me] = await Promise.all([getTranslations("ratings"), getSessionUser()]);
	const heading = (
		<h2 id="your-rating" className="font-serif text-xl font-bold">
			{t("title")}
		</h2>
	);

	let body: ReactNode;
	if (!me) {
		body = (
			<Link href={{ pathname: "/login", query: { next: path } }} className={buttonClass("primary", "w-fit")}>
				{t("loginToRate")}
			</Link>
		);
	} else if (!me.username) {
		body = (
			<Link href={{ pathname: "/onboarding", query: { next: path } }} className={buttonClass("primary", "w-fit")}>
				{t("chooseNickname")}
			</Link>
		);
	} else if (!me.emailVerified) {
		body = (
			<Notice tone="info">
				{t("verifyToRate")}{" "}
				<Link href="/verify-email" className="underline">
					{t("verifyLink")}
				</Link>
			</Notice>
		);
	} else {
		const [rating, status, lists, times] = await Promise.all([
			getUserRating(db, me.id, itemId),
			getItemStatus(db, me.id, itemId),
			getListsForItem(db, me.id, itemId),
			type === "game" ? getUserTimeToBeat(db, me.id, itemId) : null,
		]);
		body = (
			<>
				<RatingForm itemId={itemId} labels={labels} initial={rating} />
				<div className="grid gap-3 border-t border-border pt-4 sm:grid-cols-2 sm:items-start">
					<ItemStatusSelect itemId={itemId} type={type} initial={status} />
					<AddToList itemId={itemId} type={type} lists={lists} />
				</div>
				{type === "game" && status === "completed" ? <TimeToBeatForm itemId={itemId} initial={times} /> : null}
			</>
		);
	}

	return (
		<section aria-labelledby="your-rating" className="grid gap-4 rounded-lg border border-border p-4">
			{heading}
			{body}
		</section>
	);
}
