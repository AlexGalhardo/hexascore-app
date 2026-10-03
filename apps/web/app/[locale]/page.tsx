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
			{/* No visible hero by design; the page still needs one h1 for screen readers and search engines. */}
			<h1 className="sr-only">{t("title")}</h1>
			{ITEM_TYPES.map((type) => (
				<SectionHighlights key={type} type={type} locale={locale} />
			))}
		</div>
	);
}
