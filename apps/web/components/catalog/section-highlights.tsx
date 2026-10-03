import { getTranslations } from "next-intl/server";
import { CategoryLabel } from "@/components/catalog/category-label";
import { CoverThumb } from "@/components/catalog/cover-thumb";
import { ScoreBadge } from "@/components/score/score-badge";
import { Link } from "@/i18n/navigation";
import type { Locale } from "@/i18n/routing";
import { listItems } from "@/lib/catalog/cached";
import type { Sort } from "@/lib/catalog/types";
import type { ItemType } from "@/lib/db/schema/catalog";
import { itemHref, SECTION_ACCENT, SECTION_PATHS } from "@/lib/sections";

const BLOCKS = [
	{ key: "topRated", sort: "-score" },
	{ key: "newest", sort: "-release_date" },
	{ key: "mostRated", sort: "-rating_count" },
] as const satisfies readonly { key: string; sort: Sort }[];

const ROWS = 6;

// Landing block for one section: three short ranked lists side by side (stacked on mobile).
export async function SectionHighlights({ type, locale }: { type: ItemType; locale: Locale }) {
	const lists = await Promise.all(
		BLOCKS.map((b) =>
			listItems({ type, locale, sort: b.sort, limit: ROWS, ...(b.key === "topRated" ? { minScore: 0 } : {}) }),
		),
	);
	const [t, ts] = await Promise.all([getTranslations("catalog"), getTranslations("score")]);
	const accent = SECTION_ACCENT[type];
	const sectionPath = `/${SECTION_PATHS[type]}` as const;
	const title = t(`title.${type}`);

	return (
		<section aria-labelledby={`home-${type}`} className="grid gap-3">
			<div className={`flex items-baseline justify-between gap-3 border-b-2 pb-1 ${accent.border}`}>
				<h2 id={`home-${type}`} className="font-serif text-2xl font-bold">
					<Link href={sectionPath}>
						<CategoryLabel type={type} size="xl">
							{title}
						</CategoryLabel>
					</Link>
				</h2>
				<Link href={sectionPath} className={`text-sm font-medium ${accent.text}`}>
					{t("landing.seeAll", { section: title })} →
				</Link>
			</div>
			<div className="grid gap-6 md:grid-cols-3">
				{BLOCKS.map((block, i) => (
					<div key={block.key}>
						<h3 className="mb-1 text-xs font-semibold uppercase tracking-wide text-muted">
							<Link href={{ pathname: sectionPath, query: { sort: block.sort } }}>
								{t(`landing.${block.key}`)}
							</Link>
						</h3>
						{lists[i]?.data.length ? (
							<ol className="divide-y divide-border">
								{lists[i].data.map((item, rank) => (
									<li key={item.id} className="flex items-center gap-2 py-1.5">
										<span className="w-5 shrink-0 font-mono text-xs text-muted tabular-nums">
											{rank + 1}
										</span>
										<CoverThumb src={item.coverUrl} thumb={item.coverThumbUrl} />
										<Link
											href={itemHref(item.type, item.slug)}
											className="min-w-0 flex-1 truncate text-sm"
										>
											{item.title}
										</Link>
										<span className="font-mono text-xs text-muted tabular-nums">
											{block.key === "mostRated" ? item.ratingCount : item.year}
										</span>
										<ScoreBadge
											score={item.score}
											size="sm"
											label={
												item.score === null ? ts("none") : ts("outOf", { score: item.score })
											}
										/>
									</li>
								))}
							</ol>
						) : (
							<p className="py-2 text-sm text-muted">{t("landing.empty")}</p>
						)}
					</div>
				))}
			</div>
		</section>
	);
}
