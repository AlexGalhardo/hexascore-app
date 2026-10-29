import { useT } from "../../lib/useT";

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
}

const SORTS = ["popularity", "score", "release_date", "rating_count"] as const;

// A plain GET form: filters live in the URL so listings stay crawlable and shareable.
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
	const field = "min-h-11 rounded-md border border-border bg-surface px-2";

	return (
		<form method="get" action={action} className="flex flex-wrap items-end gap-3">
			<label className="grid gap-1 text-sm">
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
				<label className="grid gap-1 text-sm">
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
			<label className="grid gap-1 text-sm">
				{t("catalog.year")}
				<input
					name="year"
					type="number"
					min={1800}
					max={2100}
					defaultValue={filters.year || ""}
					className={`${field} w-24`}
				/>
			</label>
			<label className="grid gap-1 text-sm">
				{t("catalog.minScore")}
				<input
					name="min_score"
					type="number"
					min={0}
					max={10}
					step={0.5}
					defaultValue={filters.minScore ?? ""}
					className={`${field} w-24`}
				/>
			</label>
			<label className="grid gap-1 text-sm">
				{t("catalog.sort")}
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
			<button
				type="submit"
				className="min-h-11 rounded-md bg-indigo-600 px-4 font-medium text-white"
			>
				{t("catalog.apply")}
			</button>
			<a href={action} className="min-h-11 content-center text-sm underline">
				{t("catalog.reset")}
			</a>
		</form>
	);
}
