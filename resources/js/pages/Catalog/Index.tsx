import { InfiniteScroll, usePage } from "@inertiajs/react";
import Filters, { type ActiveFilters, type Facet } from "../../components/catalog/Filters";
import ItemCard from "../../components/catalog/ItemCard";
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
	facets: { genres: Facet[]; platforms: Facet[] };
	total?: number;
}

export default function Index({ type, items, filters, facets, total }: Props) {
	const t = useT();
	const { locale } = usePage<SharedProps>().props;
	const path = localized(`/${SECTION_PATHS[type]}`, locale);
	const title = t(`catalog.title.${type}`);

	return (
		<>
			<Seo title={title} description={t(`catalog.intro.${type}`)} />
			{/* Title alone, the result count on the right (design v1, round 2 item 11). */}
			<div className="flex items-end gap-6 pt-6 pb-6 md:pt-8">
				<h1 className="font-display text-5xl leading-none font-bold tracking-[-0.02em] md:text-[88px]">
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
				<InfiniteScroll data="items" as="ul" className="mt-2">
					{items.data.map((item) => (
						<ItemCard key={item.id} item={item} href={`${path}/${item.slug}`} />
					))}
				</InfiniteScroll>
			)}
			<p className="mt-6 text-sm text-muted">
				{t("community.create.missing")} <SuggestNewLink domain={type} />
			</p>
		</>
	);
}
