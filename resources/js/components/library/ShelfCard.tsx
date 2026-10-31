import { Link, usePage } from "@inertiajs/react";
import type { ReactNode } from "react";

import { localized } from "../../lib/route";
import { useT } from "../../lib/useT";
import type { SharedProps } from "../../types";
import CoverThumb from "../CoverThumb";
import { SECTION_PATHS } from "../catalog/SectionHighlights";
import type { ItemRef } from "./ItemLink";

interface Props {
	item: ItemRef;
	stars?: number | null;
	statusLabel?: string;
	children?: ReactNode;
}

// ShelfCard is the card shared by the library, list pages and the public profile tabs.
export default function ShelfCard({ item, stars, statusLabel, children }: Props) {
	const t = useT();
	const { locale } = usePage<SharedProps>().props;
	const href = localized(`/${SECTION_PATHS[item.type]}/${item.slug}`, locale);

	return (
		<article className="flex gap-3 rounded-md border border-border bg-surface p-2">
			<CoverThumb cover={item.cover} alt="" size="md" />
			<div className="min-w-0 flex-1">
				<Link href={href} className="font-medium underline">
					{item.title}
					{item.year ? (
						<span className="ml-1 font-normal text-muted">({item.year})</span>
					) : null}
				</Link>
				<p className="mt-1 flex flex-wrap gap-x-3 text-sm text-muted">
					{statusLabel && <span>{statusLabel}</span>}
					{stars !== undefined && (
						<span>{stars === null ? "—" : t("library.score", { stars })}</span>
					)}
				</p>
				{children}
			</div>
		</article>
	);
}
