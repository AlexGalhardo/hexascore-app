import type { Metadata } from "next";
import { getTranslations, setRequestLocale } from "next-intl/server";
import type { ReactNode } from "react";
import { DocsNav } from "@/components/docs/docs-nav";
import { DOC_PAGES, type DocKey } from "@/components/docs/pages";
import { Link } from "@/i18n/navigation";

type Props = { children: ReactNode; params: Promise<{ locale: string }> };

export async function generateMetadata({ params }: Props): Promise<Metadata> {
	const { locale } = await params;
	const t = await getTranslations({ locale, namespace: "api.docs" });
	return {
		title: { default: t("metaTitle"), template: `%s · ${t("metaTitle")}` },
		description: t("metaDescription"),
	};
}

export default async function DocsLayout({ children, params }: Props) {
	const { locale } = await params;
	setRequestLocale(locale);
	const t = await getTranslations("api.docs");
	const labels = Object.fromEntries(DOC_PAGES.map((p) => [p.key, t(`nav.${p.key}`)])) as Record<DocKey, string>;

	return (
		<div className="grid gap-8 lg:grid-cols-[14rem_minmax(0,1fr)]">
			<aside className="space-y-4 lg:sticky lg:top-[calc(var(--nav-h)+1rem)] lg:self-start">
				<DocsNav label={t("navLabel")} labels={labels} />
				<div className="hidden space-y-3 px-3 text-xs text-muted lg:block">
					{locale !== "en" && <p lang={locale}>{t("languageNote")}</p>}
					<Link
						href="/profile/api"
						className="inline-block font-semibold text-fg underline underline-offset-4"
					>
						{t("getKey")}
					</Link>
				</div>
			</aside>
			<article lang="en" className="min-w-0">
				{children}
			</article>
		</div>
	);
}
