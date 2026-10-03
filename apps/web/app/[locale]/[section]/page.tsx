import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { getTranslations, setRequestLocale } from "next-intl/server";
import { CatalogFilters } from "@/components/catalog/catalog-filters";
import { ItemTable } from "@/components/catalog/item-table";
import { Pagination } from "@/components/catalog/pagination";
import { asLocale } from "@/i18n/locale";
import { listGenres, listItems, listPlatforms } from "@/lib/catalog/cached";
import { InvalidCursorError } from "@/lib/catalog/cursor";
import { parseCatalogSearchParams, type RawSearchParams } from "@/lib/catalog/search-params";
import type { ItemSummary, Page } from "@/lib/catalog/types";
import { SECTION_ACCENT, SECTION_PATHS, typeFromSection } from "@/lib/sections";

type Props = {
	params: Promise<{ locale: string; section: string }>;
	searchParams: Promise<RawSearchParams>;
};

export async function generateMetadata({ params }: Props): Promise<Metadata> {
	const { locale, section } = await params;
	const type = typeFromSection(section);
	if (!type) return {};
	const t = await getTranslations({ locale, namespace: "catalog" });
	return { title: t(`title.${type}`), description: t(`intro.${type}`) };
}

export default async function CatalogPage({ params, searchParams }: Props) {
	const { locale, section } = await params;
	const type = typeFromSection(section);
	if (!type) notFound();
	setRequestLocale(locale);

	const { params: filters, query } = parseCatalogSearchParams(type, await searchParams);
	const base = { ...filters, type, locale: asLocale(locale), limit: 50 };
	let page: Page<ItemSummary>;
	try {
		page = await listItems(base);
	} catch (error) {
		// A stale or hand-edited cursor (e.g. after changing the sort) just restarts at page 1.
		if (!(error instanceof InvalidCursorError || (error as Error)?.name === "InvalidCursorError")) throw error;
		page = await listItems({ ...base, cursor: undefined });
	}
	const [genres, platforms] = await Promise.all([
		listGenres(type),
		type === "game" ? listPlatforms(type) : Promise.resolve([]),
	]);
	const t = await getTranslations("catalog");
	const accent = SECTION_ACCENT[type];

	return (
		<div className="grid gap-6">
			<header className={`border-l-4 pl-3 ${accent.border}`}>
				<h1 className="font-serif text-3xl font-bold sm:text-4xl">{t(`title.${type}`)}</h1>
				<p className="mt-1 max-w-3xl text-muted">{t(`intro.${type}`)}</p>
			</header>
			<CatalogFilters type={type} genres={genres} platforms={platforms} values={query} />
			{page.data.length ? (
				<section aria-label={t(`title.${type}`)}>
					<ItemTable items={page.data} type={type} />
					<Pagination
						pathname={`/${SECTION_PATHS[type]}`}
						query={query}
						prevCursor={page.prevCursor}
						nextCursor={page.nextCursor}
					/>
				</section>
			) : (
				<p className="py-12 text-center text-muted">{t("empty")}</p>
			)}
		</div>
	);
}
