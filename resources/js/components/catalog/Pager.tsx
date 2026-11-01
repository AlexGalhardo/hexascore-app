import { Head, Link } from "@inertiajs/react";

import { useT } from "../../lib/useT";
import { btnClass } from "../ui/Btn";

export interface Pagination {
	prev: string;
	next: string;
}

/** The URL of the page at cursor: the current path and query with `cursor` replaced ("" drops it). */
export function pageHref(url: string, cursor: string): string {
	const u = new URL(url, "http://x");
	if (cursor) u.searchParams.set("cursor", cursor);
	else u.searchParams.delete("cursor");
	return u.pathname + u.search;
}

interface Props {
	url: string;
	pagination: Pagination;
	// label names the nav landmark; without it the row is a plain group (the second copy under the list).
	label?: string;
}

// Previous/next buttons of a cursor-paged listing; a missing end renders as a disabled look-alike so the row
// keeps its size. Shown above and below the list (owner round 2026-11-01 item 12).
export default function Pager({ url, pagination, label }: Props) {
	const t = useT();
	if (!pagination.prev && !pagination.next) return null;

	const button = (cursor: string, text: string, rel: "prev" | "next") =>
		cursor ? (
			<Link href={pageHref(url, cursor)} rel={rel} className={btnClass("secondary")}>
				{text}
			</Link>
		) : (
			<span aria-disabled="true" className={btnClass("secondary")}>
				{text}
			</span>
		);

	const row = (
		<div className="flex items-center justify-between gap-4 py-4">
			{button(pagination.prev, t("catalog.previous"), "prev")}
			{button(pagination.next, t("catalog.next"), "next")}
		</div>
	);

	return label ? <nav aria-label={label}>{row}</nav> : row;
}

/** rel=prev/next head links of a paged listing, absolute against the canonical origin. */
export function PagerHead({
	url,
	pagination,
	origin,
}: {
	url: string;
	pagination: Pagination;
	origin: string;
}) {
	const abs = (cursor: string) => new URL(pageHref(url, cursor), origin).href;
	return (
		<Head>
			{pagination.prev && <link head-key="rel-prev" rel="prev" href={abs(pagination.prev)} />}
			{pagination.next && <link head-key="rel-next" rel="next" href={abs(pagination.next)} />}
		</Head>
	);
}
