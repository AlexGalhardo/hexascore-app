import { Head, Link, usePage } from "@inertiajs/react";
import CategoryIcon from "../../components/CategoryIcon";
import SuggestNewLink from "../../components/community/SuggestNewLink";
import { ACCENT_TEXT } from "../../components/home/TypeSection";
import ResultGroup, { type Hit } from "../../components/search/ResultGroup";
import { btnClass } from "../../components/ui/Btn";
import { localized } from "../../lib/route";
import { useT } from "../../lib/useT";
import type { SharedProps } from "../../types";

const PER_PAGE = 20;
const TYPES = ["game", "movie", "series", "book"] as const;
const SECTION = { game: "/games", movie: "/movies", series: "/series", book: "/books" } as const;
const NAV_KEY = {
	game: "nav.games",
	movie: "nav.movies",
	series: "nav.series",
	book: "nav.books",
} as const;

const isType = (type: string): type is (typeof TYPES)[number] =>
	(TYPES as readonly string[]).includes(type);

const textLink =
	"inline-flex min-h-11 items-center font-semibold text-primary hover:underline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fg";

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
	const href = (type: string, page: number, q = query.q) =>
		localized(
			`/search?${new URLSearchParams({ q, ...(type && { type }), ...(page > 1 && { page: String(page) }) })}`,
			locale,
		);
	const pages = Math.ceil(total / PER_PAGE);
	const num = (n: number) => new Intl.NumberFormat(locale).format(n);
	const empty =
		!tooShort && !unavailable && query.q !== "" && groups.length === 0 && results.length === 0;
	const notice = query.q === "" ? t("search.prompt") : tooShort ? t("search.tooShort") : null;

	return (
		<>
			<Head title={t("search.title")}>
				<meta name="robots" content="noindex,follow" />
			</Head>
			<h1 className="mt-6 font-display text-4xl leading-none font-bold tracking-[-0.02em] break-words md:mt-10 md:text-[64px]">
				{query.q ? t("search.resultsFor", { q: query.q }) : t("search.title")}
			</h1>
			<div className="mt-6 grid gap-x-12 gap-y-8 md:mt-8 md:grid-cols-[minmax(0,720px)_1fr]">
				<div className="min-w-0">
					<form method="get" action="/search" className="flex gap-2">
						<label className="relative flex min-h-12 flex-1 items-center border-2 border-fg bg-field-bg has-[:focus-visible]:outline-2 has-[:focus-visible]:outline-offset-2 has-[:focus-visible]:outline-primary">
							<svg
								aria-hidden="true"
								viewBox="0 0 20 20"
								fill="none"
								stroke="currentColor"
								strokeWidth="1.5"
								className="pointer-events-none absolute left-3.5 size-4 text-muted"
							>
								<circle cx="9" cy="9" r="5.5" />
								<path d="m13 13 4 4" />
							</svg>
							<span className="sr-only">{t("search.label")}</span>
							<input
								type="search"
								name="q"
								defaultValue={query.q}
								placeholder={t("nav.search")}
								className="size-full min-w-0 bg-transparent py-2 pr-3 pl-10 text-base text-fg placeholder:text-muted focus:outline-none"
							/>
						</label>
						{query.type && <input type="hidden" name="type" value={query.type} />}
						<button type="submit" className={btnClass("primary")}>
							{t("search.submit")}
						</button>
					</form>
					<nav
						aria-label={t("search.filter")}
						className="mt-6 flex flex-wrap gap-2 md:mt-8"
					>
						{["", ...TYPES].map((type) => {
							const current = query.type === type;
							return (
								<Link
									key={type}
									href={href(type, 1)}
									aria-current={current ? "page" : undefined}
									className={`inline-flex min-h-11 items-center border px-4 text-[15px] font-semibold focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fg ${current ? "border-fg bg-fg text-bg" : "border-border bg-surface hover:bg-hover"}`}
								>
									{isType(type) ? (
										<span className="inline-flex items-center gap-2">
											<CategoryIcon type={type} className="size-[18px]" />
											{t(NAV_KEY[type])}
										</span>
									) : (
										t("search.all")
									)}
								</Link>
							);
						})}
					</nav>
					{notice && (
						<p className="mt-8 border-t-2 border-fg pt-6 font-display text-2xl italic md:text-[28px]">
							{notice}
						</p>
					)}
					{unavailable && (
						<p role="status" className="mt-8 border-t-2 border-fg pt-6">
							{t("search.unavailable")}
						</p>
					)}
					{empty && (
						<div className="mt-8 border-t-2 border-fg pt-6">
							<h2 className="font-display text-[28px] leading-tight font-bold md:text-[32px]">
								{t("search.nothingFound", { q: query.q })}
							</h2>
							<p className="mt-2 text-muted">{t("search.checkSpelling")}</p>
							{didYouMean && (
								<p className="mt-4 flex flex-wrap items-baseline gap-x-3 bg-surface px-4 py-3">
									{t("search.didYouMean")}
									<Link
										href={href(query.type, 1, didYouMean)}
										className="inline-flex min-h-11 items-center font-display text-xl font-bold text-primary italic hover:underline"
									>
										{didYouMean}
									</Link>
								</p>
							)}
							<p className="mt-6 text-xs font-bold tracking-[0.1em] text-muted uppercase">
								{t("search.orBrowse")}
							</p>
							<ul className="mt-1 flex flex-wrap gap-x-5">
								{TYPES.map((type) => (
									<li key={type}>
										<Link
											href={localized(SECTION[type], locale)}
											className={`inline-flex min-h-11 items-center gap-2 font-semibold hover:underline ${ACCENT_TEXT[type]}`}
										>
											<CategoryIcon type={type} className="size-[18px]" />
											{t(NAV_KEY[type])}
										</Link>
									</li>
								))}
							</ul>
							<p className="mt-4 flex flex-wrap gap-x-4 text-sm text-muted">
								{TYPES.map((type) => (
									<SuggestNewLink key={type} domain={type} />
								))}
							</p>
						</div>
					)}
					{groups.map((g) => (
						<section key={g.type} className="mt-8 border-t-2 border-fg pt-4">
							<h2 className="flex items-baseline gap-3">
								{isType(g.type) && (
									<CategoryIcon
										type={g.type}
										className={`size-6 self-center ${ACCENT_TEXT[g.type]}`}
									/>
								)}
								<span className="font-display text-[28px] leading-tight font-bold">
									{isType(g.type) ? t(NAV_KEY[g.type]) : g.type}
								</span>
								<span className="font-sans text-sm text-muted">
									{t("search.count", { count: g.total, n: num(g.total) })}
								</span>
							</h2>
							<ResultGroup
								hits={g.items}
								searchId={searchId}
								group={g.type}
								q={query.q}
							/>
							<Link href={g.moreHref} className={`${textLink} mt-1`}>
								{t(`search.allIn.${g.type}`, { n: num(g.total), q: query.q })}
								<span aria-hidden="true">&nbsp;→</span>
							</Link>
						</section>
					))}
					{results.length > 0 && (
						<section className="mt-8 border-t-2 border-fg pt-2">
							<h2 className="sr-only">
								{t("search.count", { count: total, n: num(total) })}
							</h2>
							<ResultGroup
								hits={results}
								searchId={searchId}
								group={query.type}
								q={query.q}
								offset={(query.page - 1) * PER_PAGE}
							/>
						</section>
					)}
					{/* searchId reloads with results: the server accepts clicks only for refs of that search. */}
					{results.length > 0 && pages > 1 && (
						<nav aria-label={t("search.pagination")} className="mt-4 flex gap-6">
							{query.page > 1 && (
								<Link
									href={href(query.type, query.page - 1)}
									only={["query", "results", "total", "searchId"]}
									preserveScroll
									rel="prev"
									className={textLink}
								>
									<span aria-hidden="true">←&nbsp;</span>
									{t("search.previous")}
								</Link>
							)}
							{query.page < pages && (
								<Link
									href={href(query.type, query.page + 1)}
									only={["query", "results", "total", "searchId"]}
									preserveScroll
									rel="next"
									className={textLink}
								>
									{t("search.next")}
									<span aria-hidden="true">&nbsp;→</span>
								</Link>
							)}
						</nav>
					)}
				</div>
				{entities.length > 0 && (
					<aside className="self-start bg-surface px-5 py-6">
						<h2 className="font-sans text-xs font-bold tracking-[0.1em] uppercase">
							{t("search.people")}
						</h2>
						<ul className="mt-2">
							{entities.map((e) => (
								<li key={e.id}>
									<Link
										href={e.path}
										className="inline-flex min-h-11 items-center font-semibold text-primary underline underline-offset-4 hover:no-underline"
									>
										{e.name}
									</Link>
								</li>
							))}
						</ul>
					</aside>
				)}
			</div>
		</>
	);
}
