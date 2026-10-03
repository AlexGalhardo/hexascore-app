import type { Metadata } from "next";
import { getTranslations, setRequestLocale } from "next-intl/server";
import { z } from "zod";
import { ItemTable } from "@/components/catalog/item-table";
import { Link } from "@/i18n/navigation";
import type { Locale } from "@/i18n/routing";
import { searchItems } from "@/lib/catalog/cached";
import type { RawSearchParams } from "@/lib/catalog/search-params";
import { ITEM_TYPES } from "@/lib/db/schema/catalog";
import { SECTION_ACCENT } from "@/lib/sections";

type Props = { params: Promise<{ locale: string }>; searchParams: Promise<RawSearchParams> };

const first = (v: unknown) => (Array.isArray(v) ? v[0] : v);
const querySchema = z.object({
	q: z.preprocess(first, z.string().trim().max(100)).catch(""),
	type: z.preprocess(first, z.enum(ITEM_TYPES).optional()).catch(undefined),
});

export async function generateMetadata({ params }: Props): Promise<Metadata> {
	const { locale } = await params;
	const t = await getTranslations({ locale, namespace: "search" });
	return { title: t("title"), robots: { index: false } };
}

export default async function SearchPage({ params, searchParams }: Props) {
	const { locale } = await params;
	setRequestLocale(locale);
	const { q, type } = querySchema.parse(await searchParams);
	const [t, tcat] = await Promise.all([getTranslations("search"), getTranslations("catalog")]);
	const results =
		q.length >= 2 ? await searchItems(q.toLowerCase(), { locale: locale as Locale, type, limit: 40 }) : [];
	const groups = ITEM_TYPES.map((itemType) => ({
		type: itemType,
		items: results.filter((r) => r.type === itemType),
	})).filter((g) => g.items.length);

	return (
		<div className="grid gap-6">
			<header className="grid gap-3">
				<h1 className="font-serif text-3xl font-bold">{q ? t("resultsFor", { q }) : t("title")}</h1>
				<search>
					<form method="get" className="flex max-w-xl gap-2">
						<label htmlFor="search-page-q" className="sr-only">
							{t("label")}
						</label>
						<input
							id="search-page-q"
							name="q"
							type="search"
							defaultValue={q}
							maxLength={100}
							className="min-h-11 flex-1 rounded-md border border-border bg-bg px-3"
						/>
						{type ? <input type="hidden" name="type" value={type} /> : null}
						<button type="submit" className="min-h-11 rounded-md bg-fg px-4 text-sm font-medium text-bg">
							{t("submit")}
						</button>
					</form>
				</search>
				{q.length >= 2 ? (
					<nav aria-label={t("title")} className="flex flex-wrap gap-2 text-sm">
						{[undefined, ...ITEM_TYPES].map((itemType) => (
							<Link
								key={itemType ?? "all"}
								href={{ pathname: "/search", query: itemType ? { q, type: itemType } : { q } }}
								aria-current={itemType === type ? "page" : undefined}
								className="inline-flex min-h-11 items-center rounded-md border border-border px-3 aria-[current=page]:bg-fg aria-[current=page]:text-bg"
							>
								{itemType ? tcat(`title.${itemType}`) : t("all")}
							</Link>
						))}
					</nav>
				) : null}
			</header>

			{!q ? <p className="text-muted">{t("prompt")}</p> : null}
			{q && q.length < 2 ? <p className="text-muted">{t("tooShort")}</p> : null}
			{q.length >= 2 && !groups.length ? <p className="text-muted">{t("noResults", { q })}</p> : null}

			{groups.map((group) => (
				<section key={group.type} aria-labelledby={`results-${group.type}`} className="grid gap-2">
					<h2
						id={`results-${group.type}`}
						className={`border-l-4 pl-2 font-serif text-xl font-bold ${SECTION_ACCENT[group.type].border}`}
					>
						{tcat(`title.${group.type}`)}
					</h2>
					<ItemTable items={group.items} type={group.type} />
				</section>
			))}
		</div>
	);
}
