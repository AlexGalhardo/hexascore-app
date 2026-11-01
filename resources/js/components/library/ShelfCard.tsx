import { Link, usePage } from "@inertiajs/react";
import type { ReactNode } from "react";

import { localized } from "../../lib/route";
import { useT } from "../../lib/useT";
import type { SharedProps } from "../../types";
import CoverThumb from "../CoverThumb";
import { SECTION_PATHS } from "../catalog/SectionHighlights";
import ScoreBadge, { formatScore } from "../ui/ScoreBadge";
import type { ItemRef } from "./ItemLink";

interface Props {
	item: ItemRef;
	/** The viewer's own score; null = not rated by them, undefined = not shown. */
	stars?: number | null;
	/** Community score; undefined until the page's props carry it (FIX-library-item-score). */
	score?: number | null;
	statusLabel?: string;
	children?: ReactNode;
}

// ShelfCard is the card shared by the library, list pages and the public profile tabs (design v1).
export default function ShelfCard({ item, stars, score, statusLabel, children }: Props) {
	const t = useT();
	const { locale } = usePage<SharedProps>().props;
	const href = localized(`/${SECTION_PATHS[item.type]}/${item.slug}`, locale);

	return (
		<article className="flex h-full gap-3.5 border border-border p-3">
			<CoverThumb cover={item.cover} alt="" size="md" />
			<div className="min-w-0 flex-1">
				<Link
					href={href}
					className="font-display text-lg leading-snug font-bold hover:text-primary focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fg"
				>
					{item.title}
				</Link>
				<p className="text-sm text-muted">
					{t(`itemTypes.${item.type}`)}
					{item.year ? ` · ${item.year}` : ""}
					{statusLabel && ` · ${statusLabel}`}
				</p>
				{(score !== undefined || stars !== undefined) && (
					<p className="mt-2 flex items-center gap-2 text-sm">
						{score !== undefined && <ScoreBadge score={score} size="sm" />}
						{stars !== undefined &&
							(stars === null ? (
								<span className="text-muted">{t("library.notRated")}</span>
							) : (
								<span className="font-semibold">
									{t("library.yourScore", { stars: formatScore(stars, locale) })}
								</span>
							))}
					</p>
				)}
				{children}
			</div>
		</article>
	);
}
