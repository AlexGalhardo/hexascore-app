import { Link, usePage } from "@inertiajs/react";

import type { ItemRef } from "../../components/library/ItemLink";
import ShelfCard from "../../components/library/ShelfCard";
import ReportDialog from "../../components/ReportDialog";
import Seo from "../../components/Seo";
import ShareButton from "../../components/social/ShareButton";
import { localized } from "../../lib/route";
import { useT } from "../../lib/useT";
import type { SharedProps } from "../../types";

interface Props {
	list: {
		id: string;
		title: string;
		description: string | null;
		isPublic: boolean;
		itemCount: number;
	};
	owner: { nickname: string };
	items: { position: number; note: string | null; item: ItemRef }[];
	canEdit: boolean;
	indexable: boolean;
	/** Share-card path while the feature flag `social` is on and the page is public (SDD 14 §11 item 12). */
	ogImage?: string;
}

export default function List({ list, owner, items, canEdit, indexable, ogImage }: Props) {
	const t = useT();
	const { locale, auth } = usePage<SharedProps>().props;

	return (
		<>
			<Seo
				title={list.title}
				description={list.description ?? undefined}
				noindex={!indexable}
				image={ogImage}
			/>
			<h1 className="mt-6 font-display text-4xl leading-none font-bold tracking-[-0.02em] md:mt-10 md:text-[56px]">
				{list.title}
			</h1>
			{ogImage && <ShareButton title={list.title} />}
			<p className="mt-3 text-muted">
				<Link
					href={localized(`/u/${owner.nickname.toLowerCase()}`, locale)}
					className="underline"
				>
					{t("userLists.by", { nickname: owner.nickname })}
				</Link>
				{" · "}
				{t("userLists.itemCount", { count: list.itemCount })}
				{!list.isPublic && ` · ${t("userLists.private")}`}
			</p>
			{list.description && <p className="mt-2">{list.description}</p>}
			{canEdit && (
				<Link
					href={localized(`/lists/${list.id}/edit`, locale)}
					className="mt-2 inline-block underline"
				>
					{t("userLists.edit")}
				</Link>
			)}
			{auth?.user && !canEdit && (
				<div className="mt-2">
					<ReportDialog targetType="list" targetId={list.id} />
				</div>
			)}
			{items.length === 0 ? (
				<p className="mt-4">{t("userLists.empty")}</p>
			) : (
				<ol className="mt-8 grid gap-4 border-t-2 border-fg pt-6 sm:grid-cols-2 lg:grid-cols-3">
					{items.map((e) => (
						<li key={e.item.id}>
							<ShelfCard item={e.item}>
								{e.note && <p className="text-sm text-muted">{e.note}</p>}
							</ShelfCard>
						</li>
					))}
				</ol>
			)}
		</>
	);
}
