import { getTranslations } from "next-intl/server";
import { Link } from "@/i18n/navigation";

export default async function LocaleNotFound() {
	const t = await getTranslations("errors");
	return (
		<section className="max-w-xl py-12">
			<p className="font-mono text-sm text-muted">404</p>
			<h1 className="mt-2 font-serif text-3xl font-bold">{t("notFoundTitle")}</h1>
			<p className="mt-3 text-muted">{t("notFoundBody")}</p>
			<Link href="/" className="mt-6 inline-flex min-h-11 items-center rounded-md border border-border px-4">
				{t("home")}
			</Link>
		</section>
	);
}
