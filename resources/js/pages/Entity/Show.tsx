import { Link, usePage } from "@inertiajs/react";

import ItemCard from "../../components/catalog/ItemCard";
import { type Highlight, SECTION_PATHS } from "../../components/catalog/SectionHighlights";
import Seo, { truncate } from "../../components/Seo";
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
			<p className="text-sm text-muted">{t(`entity.kinds.${entity.kind}`)}</p>
			<h1 className="text-2xl font-bold">{entity.name}</h1>
			<Link
				href={localized(
					`/suggest/entity/${entity.id}?domain=${items.data[0]?.type ?? "game"}`,
					locale,
				)}
				className="text-sm underline"
			>
				{t("community.suggest.link")}
			</Link>
			<dl className="mt-2 flex flex-wrap gap-x-6 gap-y-1 text-sm text-muted">
				{entity.bornOn && (
					<div>
						<dt className="inline">{t("entity.born")}: </dt>
						<dd className="inline">{entity.bornOn}</dd>
					</div>
				)}
				{entity.diedOn && (
					<div>
						<dt className="inline">{t("entity.died")}: </dt>
						<dd className="inline">{entity.diedOn}</dd>
					</div>
				)}
				{entity.foundedYear !== undefined && (
					<div>
						<dt className="inline">{t("entity.founded")}: </dt>
						<dd className="inline">{entity.foundedYear}</dd>
					</div>
				)}
			</dl>
			{entity.description && (
				<p className="mt-4">
					{entity.description.text}{" "}
					<small className="text-muted">
						{entity.description.sourceUrl ? (
							<a href={entity.description.sourceUrl} rel="noopener noreferrer">
								{t("entity.descriptionSource", {
									licence: entity.description.licence,
								})}
							</a>
						) : (
							t("entity.descriptionSource", { licence: entity.description.licence })
						)}
					</small>
				</p>
			)}
			{roles.length > 0 && (
				<ul className="mt-4 flex flex-wrap gap-2 text-sm">
					{roles.map((r) => (
						<li key={r.role} className="rounded-md border border-border px-2 py-1">
							{t(`item.roles.${r.role}`)} ({r.count})
						</li>
					))}
				</ul>
			)}
			{items.data.length === 0 ? (
				<p className="mt-6">{t("entity.empty")}</p>
			) : (
				items.data.map((g) => (
					<section key={g.type} className="mt-6" aria-labelledby={`group-${g.type}`}>
						<h2 id={`group-${g.type}`} className="text-lg font-semibold">
							{t(`itemTypes.${g.type}`)}{" "}
							<span className="text-sm font-normal text-muted">
								{t("entity.showing", {
									count: g.items.length,
									total: counts[g.type] ?? g.total,
								})}
								{avgScore[g.type] !== undefined &&
									` · ${t("entity.avgScore", { score: avgScore[g.type] })}`}
							</span>
						</h2>
						<ul className="mt-2 divide-y divide-border">
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
								className="mt-2 inline-block underline"
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
