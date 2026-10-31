import { Head, Link, usePage } from "@inertiajs/react";

import ResultGroup, { type Hit } from "../../components/search/ResultGroup";
import { localized } from "../../lib/route";
import { useT } from "../../lib/useT";
import type { SharedProps } from "../../types";

const PER_PAGE = 20;
const TYPES = ["game", "movie", "series", "book"] as const;

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
	total: number;
	entities: Entity[];
	didYouMean: string | null;
	searchId: string;
	tooShort: boolean;
	unavailable?: boolean;
}

export default function Index({
	query,
	groups,
	results,
	total,
	entities,
	didYouMean,
	searchId,
	tooShort,
	unavailable = false,
}: Props) {
	const t = useT();
	const { locale } = usePage<SharedProps>().props;
	const href = (type: string, page: number) =>
		localized(
			`/search?${new URLSearchParams({ q: query.q, ...(type && { type }), ...(page > 1 && { page: String(page) }) })}`,
			locale,
		);
	const pages = Math.ceil(total / PER_PAGE);
	const empty =
		!tooShort && !unavailable && query.q !== "" && groups.length === 0 && results.length === 0;

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
			{unavailable && (
				<p role="status" className="mt-4">
					{t("search.unavailable")}
				</p>
			)}
			{empty && <p className="mt-4">{t("search.noResults", { q: query.q })}</p>}
			{empty && didYouMean && (
				<p className="mt-2">
					<Link href={`/search?q=${encodeURIComponent(didYouMean)}`}>{didYouMean}</Link>
				</p>
			)}
			{query.q !== "" && !tooShort && (
				<nav aria-label={t("search.filter")} className="mt-4 flex flex-wrap gap-1">
					{["", ...TYPES].map((type) => (
						<Link
							key={type}
							href={href(type, 1)}
							aria-current={query.type === type ? "page" : undefined}
							className={`inline-flex min-h-11 items-center rounded-md px-3 hover:underline ${query.type === type ? "bg-surface font-semibold" : ""}`}
						>
							{type ? t(`itemTypes.${type}`) : t("search.all")}
						</Link>
					))}
				</nav>
			)}
			{groups.map((g) => (
				<section key={g.type} className="mt-6">
					<h2 className="text-lg font-semibold">
						<Link href={g.moreHref} className="hover:underline">
							{t(`itemTypes.${g.type}`)}
						</Link>{" "}
						({g.total})
					</h2>
					<ResultGroup hits={g.items} searchId={searchId} group={g.type} q={query.q} />
				</section>
			))}
			{results.length > 0 && (
				<ResultGroup
					hits={results}
					searchId={searchId}
					group={query.type}
					q={query.q}
					offset={(query.page - 1) * PER_PAGE}
				/>
			)}
			{/* searchId reloads with results: the server accepts clicks only for refs of that search. */}
			{results.length > 0 && pages > 1 && (
				<nav aria-label={t("search.pagination")} className="mt-4 flex gap-4">
					{query.page > 1 && (
						<Link
							href={href(query.type, query.page - 1)}
							only={["query", "results", "total", "searchId"]}
							preserveScroll
							rel="prev"
							className="inline-flex min-h-11 items-center hover:underline"
						>
							{t("search.previous")}
						</Link>
					)}
					{query.page < pages && (
						<Link
							href={href(query.type, query.page + 1)}
							only={["query", "results", "total", "searchId"]}
							preserveScroll
							rel="next"
							className="inline-flex min-h-11 items-center hover:underline"
						>
							{t("search.next")}
						</Link>
					)}
				</nav>
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
