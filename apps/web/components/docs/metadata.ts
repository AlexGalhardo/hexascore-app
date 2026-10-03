import type { Metadata } from "next";
import { getTranslations, setRequestLocale } from "next-intl/server";
import type { DocKey } from "./pages";

export type DocPageProps = { params: Promise<{ locale: string }> };

export function docMetadata(key: DocKey): (props: DocPageProps) => Promise<Metadata> {
	return async ({ params }) => {
		const { locale } = await params;
		const t = await getTranslations({ locale, namespace: "api.docs" });
		return { title: t(`nav.${key}`) };
	};
}

export async function enterDocPage({ params }: DocPageProps): Promise<string> {
	const { locale } = await params;
	setRequestLocale(locale);
	return locale;
}
