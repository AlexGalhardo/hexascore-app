import type { Metadata } from "next";
import { setRequestLocale } from "next-intl/server";
import { LegalPage } from "@/components/legal/legal-page";
import { asLocale } from "@/i18n/locale";
import { loadLegal } from "@/lib/legal";

type Props = { params: Promise<{ locale: string }> };

export async function generateMetadata({ params }: Props): Promise<Metadata> {
	const doc = await loadLegal(asLocale((await params).locale), "policy");
	return { title: doc.title, description: doc.description };
}

export default async function PolicyPage({ params }: Props) {
	const { locale } = await params;
	setRequestLocale(locale);
	return <LegalPage doc={await loadLegal(asLocale(locale), "policy")} />;
}
