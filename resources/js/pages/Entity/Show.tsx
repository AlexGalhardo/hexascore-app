import { Link, usePage } from "@inertiajs/react";

import ItemCard from "../../components/catalog/ItemCard";
import { type Highlight, SECTION_PATHS } from "../../components/catalog/SectionHighlights";
import Seo, { truncate } from "../../components/Seo";
import { btnClass } from "../../components/ui/Btn";
import { localized, unlocalized } from "../../lib/route";
import { useT } from "../../lib/useT";
import type { SharedProps } from "../../types";

type ItemType = keyof typeof SECTION_PATHS;

interface Entity {
	id: string;
	kind: string;
	slug: string;
	code?: string;
	name: string;
	description?: { text: string; licence: string; sourceUrl?: string };
	bornOn?: string;
	diedOn?: string;
	foundedYear?: number;
}

interface Group {
	type: ItemType;
	total: number;
	items: Highlight[];
	next: number;
}

interface Props {
	entity: Entity;
	roles: { role: string; count: number }[];
	items: { data: Group[] };
	counts: Record<string, number>;
	avgScore: Record<string, number>;
	filters: { type: string };
	indexable: boolean;
}

export default function Show({
	entity,
	roles,
	items,
	counts,
	avgScore,
	filters,
	indexable,
}: Props) {
	const t = useT();
	const page = usePage<SharedProps>();
	const { locale } = page.props;
	const base = localized(unlocalized(page.url.split("?")[0]), locale);
	const pageHref = (type: string, cursor?: number) =>
		`${base}?${new URLSearchParams({ type, ...(cursor !== undefined && { cursor: String(cursor) }) })}`;

	const facts: [string, string | number | undefined][] = [
		[t("entity.born"), entity.bornOn],
		[t("entity.died"), entity.diedOn],
		[t("entity.founded"), entity.foundedYear],
	];
	const shownFacts = facts.filter(([, v]) => v !== undefined);
	const aside = shownFacts.length > 0 || entity.description;

	return (
		<>
			<Seo
				title={entity.name}
				description={truncate(
					entity.description?.text ?? t(`entity.kinds.${entity.kind}`),
					160,
				)}
				robots={indexable ? undefined : "noindex,follow"}
			/>
			<header
				className={`mt-6 grid gap-8 md:mt-10 ${aside ? "lg:grid-cols-[minmax(0,1fr)_26rem] lg:gap-x-16" : ""}`}
			>
				<div className="grid content-start justify-items-start gap-4">
					<p className="text-xs font-bold tracking-[0.1em] text-secondary uppercase">
						{t(`entity.kinds.${entity.kind}`)}
					</p>
					<h1 className="text-5xl leading-none md:text-7xl">{entity.name}</h1>
					<Link
						href={localized(
							`/suggest/entity/${entity.id}?domain=${items.data[0]?.type ?? "game"}`,
							locale,
						)}
						className={btnClass("ghost", "px-2")}
					>
						<svg
							viewBox="0 0 24 24"
							aria-hidden="true"
							className="size-4 fill-none stroke-current stroke-2"
						>
							<path d="M4 20h4L19 9l-4-4L4 16z" />
						</svg>
						{t("community.suggest.link")}
					</Link>
					{roles.length > 0 && (
						<ul className="flex flex-wrap gap-2 text-sm">
							{roles.map((r) => (
								<li
									key={r.role}
									className="rounded-xs border border-border bg-surface px-3 py-2 font-medium"
								>
									{t(`item.roles.${r.role}`)}{" "}
									<span className="text-muted tabular-nums">({r.count})</span>
								</li>
							))}
						</ul>
					)}
				</div>
				{aside && (
					<div className="grid content-start gap-4">
						{shownFacts.length > 0 && (
							<dl className="border-t-2 border-fg">
								{shownFacts.map(([k, v]) => (
									<div
										key={k}
										className="grid grid-cols-[7.5rem_minmax(0,1fr)] gap-4 border-b border-border py-2.5"
									>
										<dt className="text-muted">{k}</dt>
										<dd className="tabular-nums">{v}</dd>
									</div>
								))}
							</dl>
						)}
						{entity.description && (
							<div
								className={shownFacts.length > 0 ? "" : "border-t-2 border-fg pt-4"}
							>
								<p className="leading-relaxed">{entity.description.text}</p>
								<p className="mt-2 text-xs text-muted">
									{t("entity.descriptionSource", {
										licence: entity.description.licence,
									})}
									{entity.description.sourceUrl && (
										<>
											{" · "}
											<a
												href={entity.description.sourceUrl}
												rel="noopener noreferrer"
												className="text-primary hover:underline"
											>
												{t("entity.source")}
											</a>
										</>
									)}
								</p>
							</div>
						)}
					</div>
				)}
			</header>
			{items.data.length === 0 ? (
				<section className="mt-10 grid justify-items-center gap-3 border-t-2 border-fg pt-14 pb-6 text-center">
					<h2 className="text-3xl font-normal text-muted italic">{t("entity.empty")}</h2>
					<p className="text-sm text-muted">{t("entity.emptyHint")}</p>
					<Link href={localized("/", locale)} className={btnClass("secondary", "mt-3")}>
						{t("entity.browse")}
					</Link>
				</section>
			) : (
				items.data.map((g) => (
					<section
						key={g.type}
						className="mt-10 border-t-2 border-fg pt-4"
						aria-labelledby={`group-${g.type}`}
					>
						<h2 id={`group-${g.type}`} className="text-3xl">
							{t(`catalog.title.${g.type}`)}
							<span className="ml-3 font-sans text-sm font-normal text-muted tabular-nums">
								{t("entity.showing", {
									count: g.items.length,
									total: counts[g.type] ?? g.total,
								})}
							</span>
						</h2>
						{avgScore[g.type] !== undefined && (
							<p className="mt-1 text-sm text-muted tabular-nums">
								{t("entity.avgScore", { score: avgScore[g.type].toFixed(1) })}
							</p>
						)}
						<ul className="mt-3">
							{g.items.map((item) => (
								<ItemCard
									key={item.id}
									item={item}
									href={localized(
										`/${SECTION_PATHS[g.type]}/${item.slug}`,
										locale,
									)}
								/>
							))}
						</ul>
						{g.next >= 0 && (
							<Link
								className={btnClass("secondary", "mt-6")}
								href={
									filters.type ? pageHref(filters.type, g.next) : pageHref(g.type)
								}
							>
								{filters.type ? t("entity.next") : t("entity.seeAll")}
							</Link>
						)}
					</section>
				))
			)}
		</>
	);
}
