import { connection } from "next/server";
import { getTranslations, setRequestLocale } from "next-intl/server";
import { SectionHighlights } from "@/components/catalog/section-highlights";
import { asLocale } from "@/i18n/locale";
import { ITEM_TYPES } from "@/lib/db/schema/catalog";

type Props = { params: Promise<{ locale: string }> };

export default async function HomePage({ params }: Props) {
	const locale = asLocale((await params).locale);
	setRequestLocale(locale);
	// Render per request (data still comes from the tagged Data Cache): prerendering at build time would need a
	// reachable, migrated database in every build environment.
	await connection();
	const t = await getTranslations("home");

	return (
		<div className="grid gap-12">
			<section className="max-w-3xl">
				<h1 className="font-serif text-3xl font-bold leading-tight sm:text-5xl">{t("title")}</h1>
				<p className="mt-4 text-lg text-muted">{t("subtitle")}</p>
			</section>
			{ITEM_TYPES.map((type) => (
				<SectionHighlights key={type} type={type} locale={locale} />
			))}
		</div>
	);
}
