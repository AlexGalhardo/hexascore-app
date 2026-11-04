import { Link, usePage } from "@inertiajs/react";
import Filters, {
	type ActiveEntity,
	type ActiveFilters,
	type Facet,
	PAGE_SIZES,
} from "../../components/catalog/Filters";
import ItemTile from "../../components/catalog/ItemTile";
import Pager, { PagerHead, type Pagination, pageHref } from "../../components/catalog/Pager";
import {
	type Highlight,
	SECTION_PATHS,
	type Section,
} from "../../components/catalog/SectionHighlights";
import SuggestNewLink from "../../components/community/SuggestNewLink";
import Seo from "../../components/Seo";
import { localized } from "../../lib/route";
import { useT } from "../../lib/useT";
import type { SharedProps } from "../../types";

interface Props {
	type: Section["type"];
	items: { data: Highlight[] };
	filters: ActiveFilters;
	facets: { genres: Facet[]; platforms: Facet[]; modes?: Facet[]; watch?: Facet[] };
	active?: ActiveEntity[];
	pagination: Pagination;
	total?: number;
	/** The last page holding items (with total); a page past it shows a link back. */
	lastPage?: number;
}

export default function Index({
	type,
	items,
	filters,
	facets,
	pagination,
	total,
	lastPage,
	active = [],
}: Props) {
	const t = useT();
	const { url, props } = usePage<SharedProps>();
	const { locale, seo } = props;
	const path = localized(`/${SECTION_PATHS[type]}`, locale);
	const title = t(`catalog.title.${type}`);

	return (
		<>
			<Seo
				title={title}
				description={t(`catalog.intro.${type}`)}
				canonical={seo && pageCanonical(seo.canonical, url, filters.per)}
			/>
			{seo && <PagerHead url={url} pagination={pagination} origin={seo.canonical} />}
			<h1 className="sr-only">{title}</h1>
			{/* E-commerce layout from lg up: sticky filter sidebar left, results right (owner round 2026-11-01 item 20). */}
			<div className="lg:grid lg:grid-cols-[15rem_minmax(0,1fr)] lg:gap-10">
				<aside className="lg:sticky lg:top-24 lg:self-start">
					{total !== undefined && (
						<p className="mb-3 rounded-sm border border-border bg-surface p-4">
							<span className="block text-3xl font-bold tabular-nums">
								{new Intl.NumberFormat(locale).format(total)}
							</span>
							<span className="text-sm text-muted">
								{t(`catalog.foundLabel.${type}`)}
							</span>
						</p>
					)}
					<Filters
						action={path}
						type={type}
						filters={filters}
						genres={facets.genres}
						platforms={facets.platforms}
						watch={facets.watch}
						modes={facets.modes}
						active={active}
					/>
				</aside>
				<div className="min-w-0">
					{active.length > 0 && (
						<ul className="mb-2 flex flex-wrap gap-2">
							{active.map((a) => (
								<li key={a.param}>
									<a
										href={withoutParam(url, a.param)}
										className="inline-flex min-h-11 items-center gap-2 rounded-md border-2 border-fg bg-surface px-3 text-sm font-semibold shadow-ledge-secondary hover:bg-hover"
										aria-label={t("catalog.removeFilter", { name: a.name })}
									>
										{a.name} <span aria-hidden="true">×</span>
									</a>
								</li>
							))}
						</ul>
					)}
					{items.data.length === 0 &&
					total !== undefined &&
					total > 0 &&
					lastPage !== undefined ? (
						<p className="mt-8 text-[15px]">
							{t("catalog.pastLastPage")}{" "}
							<Link href={pageHref(url, String(lastPage))} className="underline">
								{t("catalog.lastPageLink")}
							</Link>
						</p>
					) : items.data.length === 0 ? (
						<p className="mt-8 text-[15px]">{t("catalog.empty")}</p>
					) : (
						<>
							<Pager
								url={url}
								pagination={pagination}
								label={t("catalog.pagination")}
							/>
							{/* The landing's grid card (ItemTile), so category pages and the landing match. */}
							<ul className="mt-2 grid grid-cols-2 gap-x-4 gap-y-6 sm:grid-cols-3 md:grid-cols-4 md:gap-x-5">
								{items.data.map((item) => (
									<ItemTile
										key={item.id}
										item={item}
										href={`${path}/${item.slug}`}
									/>
								))}
							</ul>
							{/* One landmark: the bottom copy repeats the buttons without a second "Pages" nav. */}
							<Pager url={url} pagination={pagination} />
						</>
					)}
				</div>
			</div>
			<p className="mt-6 text-sm text-muted">
				{t("community.create.missing")} <SuggestNewLink domain={type} />
			</p>
		</>
	);
}

/** Query keys that change a listing's result set (catalog.ParseListQuery and catalog.ListingFilters), in canonical order. */
const FILTER_PARAMS = [
	"genre",
	"platform",
	"watch",
	"mode",
	"franchise",
	"company",
	"person",
	"country",
	"language",
	"year",
	"min_score",
	"sort",
] as const;

/**
 * The canonical of a listing page: the shared canonical plus the filters that change the result set (in a fixed
 * order, other keys dropped), the page's own `page` (from 2) and a non-default `per` (the size the server applied,
 * so an invalid ?per= never leaks in). Page 2 of a filtered list is not page 2 of the unfiltered one (audit W-07).
 */
export function pageCanonical(canonical: string, url: string, per: number): string {
	const u = new URL(canonical);
	const query = new URL(url, "http://x").searchParams;
	for (const key of FILTER_PARAMS) {
		const value = query.get(key);
		if (value) u.searchParams.set(key, value);
	}
	const page = query.get("page");
	if (page && /^[1-9]\d*$/.test(page) && page !== "1") u.searchParams.set("page", page);
	if (per !== PAGE_SIZES[0]) u.searchParams.set("per", String(per));
	return u.href;
}

/** The listing URL without one filter and without the page number (the result set changes). */
export function withoutParam(url: string, param: string): string {
	const u = new URL(url, "http://x");
	u.searchParams.delete(param);
	u.searchParams.delete("page");
	return u.pathname + u.search;
}
