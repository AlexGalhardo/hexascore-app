import { InfiniteScroll, usePage } from "@inertiajs/react";
import Filters, { type ActiveFilters, type Facet } from "../../components/catalog/Filters";
import ItemCard from "../../components/catalog/ItemCard";
import {
	type Highlight,
	SECTION_PATHS,
	type Section,
} from "../../components/catalog/SectionHighlights";
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
			<h1 className="text-2xl font-bold">{title}</h1>
			<p className="mt-1 text-muted">{t(`catalog.intro.${type}`)}</p>
			<div className="mt-4">
				<Filters
					action={path}
					type={type}
					filters={filters}
					genres={facets.genres}
					platforms={facets.platforms}
				/>
			</div>
			{total !== undefined && (
				<p className="mt-4 text-sm text-muted">
					{t("catalog.resultCount", { count: total })}
				</p>
			)}
			{items.data.length === 0 ? (
				<p className="mt-4">{t("catalog.empty")}</p>
			) : (
				<InfiniteScroll data="items" as="ul" className="mt-2 divide-y divide-border">
					{items.data.map((item) => (
						<ItemCard key={item.id} item={item} href={`${path}/${item.slug}`} />
					))}
				</InfiniteScroll>
			)}
		</>
	);
}
