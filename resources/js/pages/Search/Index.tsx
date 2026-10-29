import { Head, Link } from "@inertiajs/react";

import { useT } from "../../lib/useT";

interface Hit {
	id: string;
	type: string;
	title: string;
	path: string;
	matchedTitle?: string;
}

interface Group {
	type: string;
	items: Hit[];
	total: number;
	moreHref: string;
}

interface Entity {
	id: string;
	kind: string;
	name: string;
	path: string;
}

interface Props {
	query: { q: string; type: string; locale: string; page: number };
	groups: Group[];
	results: Hit[];
	entities: Entity[];
	didYouMean: string | null;
	searchId: string;
	tooShort: boolean;
}

// Reports a click without delaying the navigation (SDD 07 §7.8).
function report(searchId: string, hit: Hit, position: number, group: string) {
	const body = JSON.stringify({ searchId, refId: hit.id, position, group });
	navigator.sendBeacon?.("/search/events", new Blob([body], { type: "application/json" }));
}

function HitList({ hits, searchId, group }: { hits: Hit[]; searchId: string; group: string }) {
	return (
		<ol className="mt-2 space-y-1">
			{hits.map((hit, i) => (
				<li key={hit.id}>
					<Link
						href={hit.path}
						onClick={() => report(searchId, hit, i + 1, group)}
						className="text-indigo-600 hover:underline dark:text-indigo-400"
					>
						{hit.title}
					</Link>
					{hit.matchedTitle && (
						<span className="ml-2 text-sm text-zinc-500">{hit.matchedTitle}</span>
					)}
				</li>
			))}
		</ol>
	);
}

// Minimal page until P1.G.3.1 styles it: props are the contract (SDD 07 §6).
export default function Index({
	query,
	groups,
	results,
	entities,
	didYouMean,
	searchId,
	tooShort,
}: Props) {
	const t = useT();
	const empty = !tooShort && query.q !== "" && groups.length === 0 && results.length === 0;

	return (
		<>
			<Head title={t("search.title")}>
				<meta name="robots" content="noindex,follow" />
			</Head>
			<h1 className="text-2xl font-semibold">
				{query.q ? t("search.resultsFor", { q: query.q }) : t("search.title")}
			</h1>
			<form method="get" action="/search" className="mt-4 flex gap-2">
				<input
					type="search"
					name="q"
					defaultValue={query.q}
					aria-label={t("search.label")}
					className="min-h-11 flex-1 rounded-md border border-zinc-300 px-3 dark:border-zinc-700 dark:bg-zinc-900"
				/>
				<button
					type="submit"
					className="min-h-11 rounded-md bg-indigo-600 px-4 font-medium text-white"
				>
					{t("search.submit")}
				</button>
			</form>
			{query.q === "" && <p className="mt-4">{t("search.prompt")}</p>}
			{tooShort && <p className="mt-4">{t("search.tooShort")}</p>}
			{empty && <p className="mt-4">{t("search.noResults", { q: query.q })}</p>}
			{empty && didYouMean && (
				<p className="mt-2">
					<Link href={`/search?q=${encodeURIComponent(didYouMean)}`}>{didYouMean}</Link>
				</p>
			)}
			{groups.map((g) => (
				<section key={g.type} className="mt-6">
					<h2 className="text-lg font-semibold">
						<Link href={g.moreHref}>{t(`itemTypes.${g.type}`)}</Link> ({g.total})
					</h2>
					<HitList hits={g.items} searchId={searchId} group={g.type} />
				</section>
			))}
			{results.length > 0 && (
				<HitList hits={results} searchId={searchId} group={query.type} />
			)}
			{entities.length > 0 && (
				<ul className="mt-6 space-y-1">
					{entities.map((e) => (
						<li key={e.id}>
							<Link href={e.path}>{e.name}</Link>
						</li>
					))}
				</ul>
			)}
		</>
	);
}
