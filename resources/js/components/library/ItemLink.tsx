import { Link, usePage } from "@inertiajs/react";
import { localized } from "../../lib/route";
import type { SharedProps } from "../../types";
import CoverThumb, { type Cover } from "../CoverThumb";
import { SECTION_PATHS, type Section } from "../catalog/SectionHighlights";

export interface ItemRef {
	id: string;
	type: Section["type"];
	slug: string;
	title: string;
	year: number | null;
	cover?: Cover | null;
}

// ItemLink points at the catalogue page of an item, locale-prefixed like the section pages.
export default function ItemLink({ item }: { item: ItemRef }) {
	const { locale } = usePage<SharedProps>().props;
	const href = localized(`/${SECTION_PATHS[item.type]}/${item.slug}`, locale);

	return (
		<Link href={href} className="inline-flex items-center gap-2 font-medium underline">
			<CoverThumb cover={item.cover} alt="" />
			<span>
				{item.title}
				{item.year ? (
					<span className="ml-1 font-normal text-muted">({item.year})</span>
				) : null}
			</span>
		</Link>
	);
}
