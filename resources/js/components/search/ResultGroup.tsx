import { Link } from "@inertiajs/react";
import type { ReactNode } from "react";

import { useT } from "../../lib/useT";
import CoverThumb, { type Cover } from "../CoverThumb";
import ScoreBadge from "../ui/ScoreBadge";

export interface Hit {
	id: string;
	type: string;
	title: string;
	path: string;
	matchedTitle?: string;
	cover?: Cover | null;
	// Not sent by the search controller yet (FIX-search-hit-meta); the row shows them once it does.
	year?: number | null;
	score?: number | null;
}

// Reports a click without delaying the navigation (SDD 07 §7.8). Only hits the
// page rendered are reported: the server checks them against the stored result refs.
export function reportClick(searchId: string, hit: Hit, position: number, group: string): void {
	const body = JSON.stringify({ searchId, refId: hit.id, position, group });
	navigator.sendBeacon?.("/search/events", new Blob([body], { type: "application/json" }));
}

export function highlight(text: string, q: string): ReactNode {
	const needle = q.trim();
	if (needle.length < 2) return text;
	const at = text.toLocaleLowerCase().indexOf(needle.toLocaleLowerCase());
	if (at < 0) return text;
	return (
		<>
			{text.slice(0, at)}
			<mark className="bg-transparent font-semibold text-inherit">
				{text.slice(at, at + needle.length)}
			</mark>
			{text.slice(at + needle.length)}
		</>
	);
}

interface Props {
	hits: Hit[];
	searchId: string;
	group: string;
	q: string;
	// Positions are 1-based within the group; a later page continues the count.
	offset?: number;
}

export default function ResultGroup({ hits, searchId, group, q, offset = 0 }: Props) {
	const t = useT();

	return (
		<ol>
			{hits.map((hit, i) => (
				<li key={hit.id} className="flex items-center gap-4 border-b border-border py-3">
					<CoverThumb cover={hit.cover} alt="" size="md" />
					<div className="min-w-0 flex-1">
						<Link
							href={hit.path}
							onClick={() => reportClick(searchId, hit, offset + i + 1, group)}
							className="font-display text-lg leading-snug font-bold hover:text-primary focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fg"
						>
							{highlight(hit.title, q)}
						</Link>
						<p className="text-sm text-muted">
							{t(`itemTypes.${hit.type}`)}
							{hit.year ? ` · ${hit.year}` : ""}
							{hit.matchedTitle && ` · ${hit.matchedTitle}`}
						</p>
					</div>
					{hit.score !== undefined && <ScoreBadge score={hit.score} size="sm" />}
				</li>
			))}
		</ol>
	);
}
