import { Head, Link } from "@inertiajs/react";

import { useT } from "../../lib/useT";
import { btnClass } from "../ui/Btn";

/** Page numbers of the previous and next pages ("" at either end). */
export interface Pagination {
	prev: string;
	next: string;
}

/**
 * The URL of page number `page`: the current path and query with `page` replaced. Page 1 drops it, and an old
 * `cursor` from before numbered pages is dropped too.
 */
export function pageHref(url: string, page: string): string {
	const u = new URL(url, "http://x");
	u.searchParams.delete("cursor");
	if (page && page !== "1") u.searchParams.set("page", page);
	else u.searchParams.delete("page");
	return u.pathname + u.search;
}

interface Props {
	url: string;
	pagination: Pagination;
	// label names the nav landmark; without it the row is a plain group (the second copy under the list).
	label?: string;
}

// Previous/next buttons of a numbered listing; a missing end renders as a disabled look-alike so the row
// keeps its size. Shown above and below the list (owner round 2026-11-01 item 12).
export default function Pager({ url, pagination, label }: Props) {
	const t = useT();
	if (!pagination.prev && !pagination.next) return null;

	const button = (page: string, text: string, rel: "prev" | "next") =>
		page ? (
			<Link href={pageHref(url, page)} rel={rel} className={btnClass("secondary")}>
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
	const abs = (page: string) => new URL(pageHref(url, page), origin).href;
	return (
		<Head>
			{pagination.prev && <link head-key="rel-prev" rel="prev" href={abs(pagination.prev)} />}
			{pagination.next && <link head-key="rel-next" rel="next" href={abs(pagination.next)} />}
		</Head>
	);
}
