import { Link } from "@inertiajs/react";
import type { ReactNode } from "react";

export interface Hit {
	id: string;
	type: string;
	title: string;
	path: string;
	matchedTitle?: string;
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
	return (
		<ol className="mt-2 space-y-1">
			{hits.map((hit, i) => (
				<li key={hit.id}>
					<Link
						href={hit.path}
						onClick={() => reportClick(searchId, hit, offset + i + 1, group)}
						className="inline-flex min-h-11 items-center text-accent-game hover:underline"
					>
						{highlight(hit.title, q)}
					</Link>
					{hit.matchedTitle && (
						<span className="ml-2 text-sm text-muted">{hit.matchedTitle}</span>
					)}
				</li>
			))}
		</ol>
	);
}
