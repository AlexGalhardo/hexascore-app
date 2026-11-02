import { usePage } from "@inertiajs/react";
import Filters, {
	type ActiveEntity,
	type ActiveFilters,
	type Facet,
	PAGE_SIZES,
} from "../../components/catalog/Filters";
import ItemTile from "../../components/catalog/ItemTile";
import Pager, { PagerHead, type Pagination } from "../../components/catalog/Pager";
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
	facets: { genres: Facet[]; platforms: Facet[]; modes?: Facet[] };
	active?: ActiveEntity[];
	pagination: Pagination;
	total?: number;
}

export default function Index({
	type,
	items,
	filters,
	facets,
	pagination,
	total,
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
			{/* Title alone, the result count on the right (design v1, round 2 item 11). */}
			<div className="flex items-end gap-6 pb-6">
				<h1 className="font-display text-4xl leading-none font-bold tracking-[-0.02em] md:text-6xl">
					{title}
				</h1>
				{total !== undefined && (
					<p className="ml-auto hidden flex-col items-end md:flex">
						<span className="text-4xl leading-none font-bold tabular-nums">
							{new Intl.NumberFormat(locale).format(total)}
						</span>
						<span className="mt-1 text-[13px] text-muted">
							{t(`catalog.foundLabel.${type}`)}
						</span>
					</p>
				)}
			</div>
			{/* E-commerce layout from lg up: sticky filter sidebar left, results right (owner round 2026-11-01 item 20). */}
			<div className="lg:grid lg:grid-cols-[15rem_minmax(0,1fr)] lg:gap-10">
				<aside className="lg:sticky lg:top-24 lg:self-start">
					<Filters
						action={path}
						type={type}
						filters={filters}
						genres={facets.genres}
						platforms={facets.platforms}
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
										className="inline-flex min-h-11 items-center gap-2 rounded-xs border-2 border-fg bg-surface px-3 text-sm font-semibold shadow-ledge-secondary hover:bg-fg hover:text-bg"
										aria-label={t("catalog.removeFilter", { name: a.name })}
									>
										{a.name} <span aria-hidden="true">×</span>
									</a>
								</li>
							))}
						</ul>
					)}
					{total !== undefined && (
						<p className="mt-4 text-sm text-muted md:sr-only">
							{new Intl.NumberFormat(locale).format(total)}{" "}
							{t(`catalog.foundLabel.${type}`)}
						</p>
					)}
					{items.data.length === 0 ? (
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

/**
 * The canonical of a listing page: the shared canonical plus the page's own `page` (from 2) and a non-default
 * `per` (the size the server applied, so an invalid ?per= never leaks in); filters stay out.
 */
export function pageCanonical(canonical: string, url: string, per: number): string {
	const u = new URL(canonical);
	const page = new URL(url, "http://x").searchParams.get("page");
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
