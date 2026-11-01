import { Link, usePage } from "@inertiajs/react";
import { localized } from "../../lib/route";
import { useT } from "../../lib/useT";
import type { SharedProps } from "../../types";
import CoverThumb from "../CoverThumb";
import { type Highlight, SECTION_PATHS, type Section } from "./SectionHighlights";

export interface ForYouReason {
	kind: "rated" | "completed" | "wanted";
	itemId: string;
	title: string;
	stars?: number;
}

export interface ForYouList {
	basis: "personal" | "onboarding";
	data: { item: Highlight; reason: ForYouReason | null }[];
}

// One recommendation list; the reason is worded here so it follows the catalogue locale.
export default function ForYouListView({
	type,
	list,
}: {
	type: Section["type"];
	list: ForYouList;
}) {
	const t = useT();
	const { locale } = usePage<SharedProps>().props;
	const path = localized(`/${SECTION_PATHS[type]}`, locale);

	return (
		<ol className="divide-y divide-border">
			{list.data.map(({ item, reason }) => (
				<li key={item.id} className="flex items-center gap-3 py-2">
					<CoverThumb cover={item.cover} alt="" />
					<div className="min-w-0 flex-1">
						<Link
							href={`${path}/${item.slug}?from=foryou`}
							className="block truncate font-medium"
						>
							{item.title}
						</Link>
						{reason && (
							<span className="text-xs text-muted">
								{t(`forYou.${reason.kind}`, {
									title: reason.title,
									stars: reason.stars ?? 0,
								})}
							</span>
						)}
					</div>
				</li>
			))}
		</ol>
	);
}
