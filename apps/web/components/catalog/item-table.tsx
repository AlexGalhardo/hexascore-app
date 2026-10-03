import { getTranslations } from "next-intl/server";
import { ScoreBadge } from "@/components/score/score-badge";
import { Link } from "@/i18n/navigation";
import type { ItemSummary } from "@/lib/catalog/types";
import { abbreviateLabel, CATEGORY_KEYS } from "@/lib/categories";
import type { ItemType } from "@/lib/db/schema/catalog";
import { itemHref } from "@/lib/sections";

type Props = {
	items: ItemSummary[];
	/** Single-type lists show the 6 category columns on wide screens; mixed lists show the type instead. */
	type?: ItemType;
	/** Offset for the rank column (null hides it). */
	rankFrom?: number | null;
};

// Dense table from md up, stacked list below (design-system "data table").
export async function ItemTable({ items, type, rankFrom = null }: Props) {
	const t = await getTranslations("catalog");
	const tc = await getTranslations("categories");
	const ts = await getTranslations("score");
	const tt = await getTranslations("itemTypes");
	const categoryLabels = type ? CATEGORY_KEYS[type].map((k) => tc(`${type}.${k}`)) : [];
	const scoreLabel = (score: number | null) => (score === null ? ts("none") : ts("outOf", { score }));

	return (
		<>
			<table className="hidden w-full border-collapse text-sm md:table">
				<thead className="sticky top-0 bg-bg">
					<tr className="border-b border-border text-left text-xs uppercase tracking-wide text-muted">
						{rankFrom !== null ? <th className="w-10 py-2 pr-2 font-medium">#</th> : null}
						<th className="py-2 pr-4 font-medium">{t("columns.title")}</th>
						<th className="w-16 py-2 pr-4 font-medium">{t("columns.year")}</th>
						<th className="w-16 py-2 pr-4 font-medium">{t("columns.score")}</th>
						{categoryLabels.map((label) => (
							<th key={label} className="hidden w-14 py-2 pr-2 text-right font-medium lg:table-cell">
								<abbr title={label} className="no-underline">
									{abbreviateLabel(label)}
								</abbr>
							</th>
						))}
						<th className="w-24 py-2 text-right font-medium">{t("columns.ratings")}</th>
					</tr>
				</thead>
				<tbody>
					{items.map((item, i) => (
						<tr key={item.id} className="border-b border-border align-middle hover:bg-surface">
							{rankFrom !== null ? (
								<td className="py-2 pr-2 font-mono text-muted tabular-nums">{rankFrom + i}</td>
							) : null}
							<td className="py-2 pr-4">
								<Link href={itemHref(item.type, item.slug)} className="font-medium hover:underline">
									{item.title}
								</Link>
								<span className="block truncate text-xs text-muted">
									{type ? item.genres.join(", ") : `${tt(item.type)} · ${item.genres.join(", ")}`}
								</span>
							</td>
							<td className="py-2 pr-4 font-mono tabular-nums">{item.year ?? "—"}</td>
							<td className="py-2 pr-4">
								<ScoreBadge score={item.score} label={scoreLabel(item.score)} size="sm" />
							</td>
							{categoryLabels.map((label, c) => (
								<td
									key={label}
									className="hidden py-2 pr-2 text-right font-mono tabular-nums text-muted lg:table-cell"
								>
									{item.categories[c] ?? "–"}
								</td>
							))}
							<td className="py-2 text-right font-mono text-muted tabular-nums">{item.ratingCount}</td>
						</tr>
					))}
				</tbody>
			</table>

			<ul className="divide-y divide-border border-y border-border md:hidden">
				{items.map((item, i) => (
					<li key={item.id} className="flex items-center gap-3 py-3">
						{rankFrom !== null ? (
							<span className="w-6 shrink-0 font-mono text-sm text-muted tabular-nums">
								{rankFrom + i}
							</span>
						) : null}
						<div className="min-w-0 flex-1">
							<Link href={itemHref(item.type, item.slug)} className="font-medium">
								{item.title}
							</Link>
							<p className="truncate text-xs text-muted">
								{[type ? null : tt(item.type), item.year, item.genres.join(", ")]
									.filter(Boolean)
									.join(" · ")}
							</p>
							<p className="text-xs text-muted">{t("ratingCount", { count: item.ratingCount })}</p>
						</div>
						<ScoreBadge score={item.score} label={scoreLabel(item.score)} size="md" />
					</li>
				))}
			</ul>
		</>
	);
}
