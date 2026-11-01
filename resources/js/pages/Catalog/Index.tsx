import { usePage } from "@inertiajs/react";
import CategoryIcon from "../../components/CategoryIcon";
import Filters, { type ActiveFilters, type Facet } from "../../components/catalog/Filters";
import ItemCard from "../../components/catalog/ItemCard";
import Pager, { PagerHead, type Pagination } from "../../components/catalog/Pager";
import {
	type Highlight,
	SECTION_PATHS,
	type Section,
} from "../../components/catalog/SectionHighlights";
import SuggestNewLink from "../../components/community/SuggestNewLink";
import { ACCENT_TEXT } from "../../components/home/TypeSection";
import Seo from "../../components/Seo";
import { localized } from "../../lib/route";
import { useT } from "../../lib/useT";
import type { SharedProps } from "../../types";

interface Props {
	type: Section["type"];
	items: { data: Highlight[] };
	filters: ActiveFilters;
	facets: { genres: Facet[]; platforms: Facet[] };
	pagination: Pagination;
	total?: number;
}

export default function Index({ type, items, filters, facets, pagination, total }: Props) {
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
				canonical={seo && pageCanonical(seo.canonical, url)}
			/>
			{seo && <PagerHead url={url} pagination={pagination} origin={seo.canonical} />}
			{/* Title alone, the result count on the right (design v1, round 2 item 11). */}
			<div className="flex items-end gap-6 pt-6 pb-6 md:pt-8">
				<h1 className="flex items-center gap-3 font-display text-5xl leading-none font-bold tracking-[-0.02em] md:gap-5 md:text-[88px]">
					<CategoryIcon
						type={type}
						className={`size-10 md:size-16 ${ACCENT_TEXT[type]}`}
					/>
					{title}
				</h1>
				{total !== undefined && (
					<p className="ml-auto hidden flex-col items-end md:flex">
						<span className="text-5xl leading-none font-bold tabular-nums">
							{new Intl.NumberFormat(locale).format(total)}
						</span>
						<span className="mt-1 text-[13px] text-muted">
							{t("catalog.itemsFoundLabel", { count: total })}
						</span>
					</p>
				)}
			</div>
			<Filters
				action={path}
				type={type}
				filters={filters}
				genres={facets.genres}
				platforms={facets.platforms}
			/>
			{total !== undefined && (
				<p className="mt-4 text-sm text-muted md:sr-only">
					{t("catalog.resultCount", { count: total })}
				</p>
			)}
			{items.data.length === 0 ? (
				<p className="mt-8 text-[15px]">{t("catalog.empty")}</p>
			) : (
				<>
					<Pager url={url} pagination={pagination} label={t("catalog.pagination")} />
					<ul className="mt-2">
						{items.data.map((item) => (
							<ItemCard key={item.id} item={item} href={`${path}/${item.slug}`} />
						))}
					</ul>
					{/* One landmark: the bottom copy repeats the buttons without a second "Pages" nav. */}
					<Pager url={url} pagination={pagination} />
				</>
			)}
			<p className="mt-6 text-sm text-muted">
				{t("community.create.missing")} <SuggestNewLink domain={type} />
			</p>
		</>
	);
}

/** The canonical of a listing page: the shared canonical plus the page's own `cursor`; filters stay out. */
export function pageCanonical(canonical: string, url: string): string {
	const cursor = new URL(url, "http://x").searchParams.get("cursor");
	const u = new URL(canonical);
	if (cursor) u.searchParams.set("cursor", cursor);
	return u.href;
}
