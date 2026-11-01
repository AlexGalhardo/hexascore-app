import { Link, usePage } from "@inertiajs/react";
import type { ReactNode } from "react";

import { localized } from "../../lib/route";
import { useT } from "../../lib/useT";
import type { SharedProps } from "../../types";
import CoverThumb from "../CoverThumb";
import { type Highlight, SECTION_PATHS, type Section } from "../catalog/SectionHighlights";
import ScoreBadge from "../ui/ScoreBadge";

// Item page props from app/recommend/related (P6.D.1.3); both lists are always arrays.
export interface SimilarItem extends Highlight {
	reason: { kind: string; entity: string } | null;
}

export interface CrossItem extends Highlight {
	relation: "based_on" | "adapted_as";
}

interface Props {
	type: Section["type"];
	similar: SimilarItem[];
	crossMedia: CrossItem[];
}

// "More games like this" and the cross-media strip of design v1: horizontal rows of 2:3 cover cards.
export default function MoreLikeThis({ type, similar, crossMedia }: Props) {
	const t = useT();

	return (
		<>
			{similar.length > 0 && (
				<Strip id="similar" title={t(`item.moreLike.${type}`)}>
					{similar.map((it) => (
						<Card
							key={it.id}
							item={it}
							from="similar"
							note={
								it.reason
									? t("item.similarReason", { entity: it.reason.entity })
									: null
							}
						/>
					))}
				</Strip>
			)}
			{crossMedia.length > 0 && (
				<Strip id="crossmedia" title={t("item.crossMedia")}>
					{crossMedia.map((it) => (
						<Card
							key={it.id}
							item={it}
							from="crossmedia"
							note={`${t(`itemTypes.${it.type}`)} · ${t(`item.relation.${it.relation}`)}`}
						/>
					))}
				</Strip>
			)}
		</>
	);
}

function Strip({ id, title, children }: { id: string; title: string; children: ReactNode }) {
	return (
		<section
			aria-labelledby={`${id}-heading`}
			className="grid content-start gap-4 border-t-2 border-fg pt-4"
		>
			<h2 id={`${id}-heading`} className="text-2xl">
				{title}
			</h2>
			<ul className="-mx-1 flex snap-x gap-4 overflow-x-auto px-1 pb-2 md:gap-5">
				{children}
			</ul>
		</section>
	);
}

function Card({
	item,
	from,
	note,
}: {
	item: Highlight;
	from: "similar" | "crossmedia";
	note: string | null;
}) {
	const { locale } = usePage<SharedProps>().props;
	const section = SECTION_PATHS[item.type as Section["type"]] ?? "games";
	// ?from= counts the click in the related-block metrics (app/recommend/related).
	const href = `${localized(`/${section}/${item.slug}`, locale)}?from=${from}`;

	return (
		<li className="grid w-30 shrink-0 snap-start content-start gap-2 md:w-32">
			<Link
				href={href}
				className="group grid gap-2 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fg"
			>
				<CoverThumb cover={item.cover} alt="" size="card" />
				<span className="line-clamp-2 text-sm leading-snug font-semibold group-hover:text-primary">
					{item.title}
				</span>
			</Link>
			<div className="flex items-center justify-between gap-2">
				<span className="text-[13px] text-muted tabular-nums">{item.year}</span>
				<ScoreBadge score={item.score} size="xs" />
			</div>
			{note && <p className="text-xs text-muted">{note}</p>}
		</li>
	);
}
