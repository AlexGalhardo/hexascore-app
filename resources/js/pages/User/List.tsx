import { Link, usePage } from "@inertiajs/react";

import type { ItemRef } from "../../components/library/ItemLink";
import ShelfCard from "../../components/library/ShelfCard";
import ReportDialog from "../../components/ReportDialog";
import Seo from "../../components/Seo";
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
}

export default function List({ list, owner, items, canEdit, indexable }: Props) {
	const t = useT();
	const { locale, auth } = usePage<SharedProps>().props;

	return (
		<>
			<Seo
				title={list.title}
				description={list.description ?? undefined}
				noindex={!indexable}
			/>
			<h1 className="text-2xl font-semibold">{list.title}</h1>
			<p className="mt-1 text-muted">
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
				<ol className="mt-4 divide-y divide-border">
					{items.map((e) => (
						<li key={e.item.id} className="py-2">
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
