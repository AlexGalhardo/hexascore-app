import { useId, useState } from "react";

import { useT } from "../../lib/useT";
import { btnClass } from "../ui/Btn";

export interface Facet {
	slug: string;
	name: string;
	count: number;
}

export interface ActiveFilters {
	genre: string;
	platform: string;
	year: number;
	minScore: number | null;
	sort: string;
	per: number;
}

const SORTS = ["popularity", "score", "release_date", "rating_count"] as const;

/** The page sizes the server honours for ?per= (catalog_controller.go); the first is the default. */
export const PAGE_SIZES = [25, 50, 75, 100] as const;

/** Filters other than sort that differ from their default. */
export function activeCount(f: ActiveFilters): number {
	return [f.genre, f.platform, f.year, f.minScore].filter(
		(v) => v !== "" && v !== 0 && v !== null,
	).length;
}

// A plain GET form: filters and the page size live in the URL so listings stay crawlable and shareable. From
// lg up it is the left sidebar of the listing (sticky, owner round 2026-11-01 item 20); below lg the fields
// collapse behind the "Filters · N" disclosure button and push the list down (no overlay); sort stays visible.
export default function Filters({
	action,
	type,
	filters,
	genres,
	platforms,
}: {
	action: string;
	type: string;
	filters: ActiveFilters;
	genres: Facet[];
	platforms: Facet[];
}) {
	const t = useT();
	const [open, setOpen] = useState(false);
	const panel = useId();
	const count = activeCount(filters);
	// Full-width fields: a select sized to its longest option used to shift the list when facets arrived (CLS).
	const field =
		"min-h-11 w-full min-w-0 rounded-xs border border-field bg-field-bg px-3 text-[15px] text-fg focus-visible:outline-2 focus-visible:outline-offset-1 focus-visible:outline-primary";
	const label = "grid gap-1.5 text-[13px] font-semibold";

	return (
		<form
			method="get"
			action={action}
			className="grid grid-cols-[auto_minmax(0,1fr)] items-end gap-2 border-t-2 border-b border-fg py-3 lg:grid-cols-1 lg:gap-4 lg:border-b-0 lg:py-4"
		>
			<button
				type="button"
				aria-expanded={open}
				aria-controls={panel}
				onClick={() => setOpen((o) => !o)}
				className={btnClass(
					"secondary",
					"px-3.5 aria-expanded:bg-fg aria-expanded:text-bg aria-expanded:hover:bg-fg lg:hidden",
				)}
			>
				<svg aria-hidden="true" viewBox="0 0 24 24" className="size-4" fill="none">
					<path
						d={open ? "M6 6 L18 18 M18 6 L6 18" : "M4 6 H20 M7 12 H17 M10 18 H14"}
						stroke="currentColor"
						strokeWidth="1.75"
						strokeLinecap="round"
					/>
				</svg>
				{t("catalog.filters", { count })}
			</button>
			<label className="grid min-w-0 gap-1.5 text-[13px] font-semibold">
				<span className="sr-only lg:not-sr-only">{t("catalog.sort")}</span>
				<select name="sort" defaultValue={filters.sort} className={field}>
					{SORTS.flatMap((s) => [`-${s}`, s]).map((s) => (
						<option key={s} value={s}>
							{t(
								`catalog.sorts.${s.replace(/^-/, "")}_${s.startsWith("-") ? "desc" : "asc"}`,
							)}
						</option>
					))}
				</select>
			</label>
			<div
				id={panel}
				className={`${open ? "grid" : "hidden"} col-span-2 gap-4 bg-surface p-4 lg:col-span-1 lg:grid lg:bg-transparent lg:p-0`}
			>
				<label className={label}>
					{t("catalog.perPage")}
					<select name="per" defaultValue={filters.per} className={field}>
						{PAGE_SIZES.map((n) => (
							<option key={n} value={n}>
								{n}
							</option>
						))}
					</select>
				</label>
				<label className={label}>
					{t("catalog.genre")}
					<select name="genre" defaultValue={filters.genre} className={field}>
						<option value="">{t("catalog.anyGenre")}</option>
						{genres.map((g) => (
							<option key={g.slug} value={g.slug}>
								{g.name} ({g.count})
							</option>
						))}
					</select>
				</label>
				{type === "game" && (
					<label className={label}>
						{t("catalog.platform")}
						<select name="platform" defaultValue={filters.platform} className={field}>
							<option value="">{t("catalog.anyPlatform")}</option>
							{platforms.map((p) => (
								<option key={p.slug} value={p.slug}>
									{p.name} ({p.count})
								</option>
							))}
						</select>
					</label>
				)}
				<label className={label}>
					{t("catalog.year")}
					<input
						name="year"
						type="number"
						min={1800}
						max={2100}
						defaultValue={filters.year || ""}
						className={field}
					/>
				</label>
				<label className={label}>
					{t("catalog.minScore")}
					<input
						name="min_score"
						type="number"
						min={0}
						max={10}
						step={0.5}
						defaultValue={filters.minScore ?? ""}
						className={field}
					/>
				</label>
				<div className="grid grid-cols-2 gap-3">
					<button type="submit" className={btnClass("primary")}>
						{t("catalog.apply")}
					</button>
					<a href={action} className={btnClass("secondary")}>
						{t("catalog.reset")}
					</a>
				</div>
			</div>
		</form>
	);
}
