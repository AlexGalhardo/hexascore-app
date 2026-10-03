import { getTranslations } from "next-intl/server";
import { Link } from "@/i18n/navigation";
import { sortMessageKey } from "@/lib/catalog/search-params";
import { SORTS } from "@/lib/catalog/types";
import { CATEGORY_KEYS } from "@/lib/categories";
import type { ItemType } from "@/lib/db/schema/catalog";
import { SECTION_ACCENT, SECTION_PATHS } from "@/lib/sections";

type Props = {
	type: ItemType;
	genres: string[];
	platforms: string[];
	values: Record<string, string>;
};

const field = "min-h-11 w-full rounded-md border border-border bg-bg px-2 text-sm";
const label = "flex flex-col gap-1 text-xs font-medium text-muted";

// Plain GET form: filters live in the URL and work without JavaScript. Submitting drops the cursor (page 1).
export async function CatalogFilters({ type, genres, platforms, values }: Props) {
	const t = await getTranslations("catalog");
	const tc = await getTranslations("categories");
	const accent = SECTION_ACCENT[type];
	const anyCategory = CATEGORY_KEYS[type].some((k) => values[`min_${k}`]);

	return (
		<form method="get" className="grid gap-3 rounded-md border border-border p-3" aria-label={t("filters")}>
			<div className="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-6">
				<label className={label}>
					{t("genre")}
					<select name="genre" defaultValue={values.genre ?? ""} className={field}>
						<option value="">{t("anyGenre")}</option>
						{genres.map((g) => (
							<option key={g} value={g}>
								{g}
							</option>
						))}
					</select>
				</label>
				{type === "game" ? (
					<label className={label}>
						{t("platform")}
						<select name="platform" defaultValue={values.platform ?? ""} className={field}>
							<option value="">{t("anyPlatform")}</option>
							{platforms.map((p) => (
								<option key={p} value={p}>
									{p}
								</option>
							))}
						</select>
					</label>
				) : null}
				<label className={label}>
					{t("year")}
					<input
						name="year"
						type="number"
						inputMode="numeric"
						min={1800}
						max={2100}
						defaultValue={values.year}
						className={field}
					/>
				</label>
				<label className={label}>
					{t("minScore")}
					<input
						name="min_score"
						type="number"
						inputMode="numeric"
						min={0}
						max={100}
						defaultValue={values.min_score}
						className={field}
					/>
				</label>
				<label className={`${label} col-span-2 sm:col-span-1`}>
					{t("sort")}
					<select name="sort" defaultValue={values.sort ?? SORTS[0]} className={field}>
						{SORTS.map((s) => (
							<option key={s} value={s}>
								{t(`sorts.${sortMessageKey(s)}`)}
							</option>
						))}
					</select>
				</label>
			</div>
			<details open={anyCategory} className="group">
				<summary className="min-h-11 cursor-pointer content-center text-sm font-medium">
					{t("minCategories")}
				</summary>
				<div className="grid grid-cols-2 gap-3 pt-2 sm:grid-cols-3 lg:grid-cols-6">
					{CATEGORY_KEYS[type].map((key) => (
						<label key={key} className={label}>
							{tc(`${type}.${key}`)}
							<input
								name={`min_${key}`}
								type="number"
								inputMode="numeric"
								min={0}
								max={100}
								defaultValue={values[`min_${key}`]}
								className={field}
							/>
						</label>
					))}
				</div>
			</details>
			<div className="flex gap-2">
				<button type="submit" className={`min-h-11 rounded-md px-4 text-sm font-semibold text-bg ${accent.bg}`}>
					{t("apply")}
				</button>
				<Link
					href={`/${SECTION_PATHS[type]}`}
					className="inline-flex min-h-11 items-center rounded-md border border-border px-4 text-sm"
				>
					{t("reset")}
				</Link>
			</div>
		</form>
	);
}
