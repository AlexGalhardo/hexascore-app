import { getTranslations, setRequestLocale } from "next-intl/server";

type Props = { params: Promise<{ locale: string }> };

export default async function HomePage({ params }: Props) {
	const { locale } = await params;
	setRequestLocale(locale);
	const t = await getTranslations("home");

	return (
		<section className="max-w-3xl">
			<h1 className="font-serif text-3xl font-bold leading-tight sm:text-5xl">{t("title")}</h1>
			<p className="mt-4 text-lg text-muted">{t("subtitle")}</p>
		</section>
	);
}
