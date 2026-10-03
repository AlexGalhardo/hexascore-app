import type { Metadata } from "next";
import { setRequestLocale } from "next-intl/server";
import { LegalPage } from "@/components/legal/legal-page";
import type { Locale } from "@/i18n/routing";
import { loadLegal } from "@/lib/legal";

type Props = { params: Promise<{ locale: string }> };

export async function generateMetadata({ params }: Props): Promise<Metadata> {
	const doc = await loadLegal((await params).locale as Locale, "terms");
	return { title: doc.title, description: doc.description };
}

export default async function TermsPage({ params }: Props) {
	const { locale } = await params;
	setRequestLocale(locale);
	return <LegalPage doc={await loadLegal(locale as Locale, "terms")} />;
}
