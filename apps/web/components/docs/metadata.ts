import type { Metadata } from "next";
import { getTranslations, setRequestLocale } from "next-intl/server";
import type { DocKey } from "./pages";

export type DocPageProps = { params: Promise<{ locale: string }> };

export function docMetadata(key: DocKey): (props: DocPageProps) => Promise<Metadata> {
	return async ({ params }) => {
		const { locale } = await params;
		const [nav, docs] = await Promise.all([
			getTranslations({ locale, namespace: "api.docs" }),
			getTranslations({ locale, namespace: "docs" }),
		]);
		return { title: nav(`nav.${key}`), description: docs(`${key}.lead`) };
	};
}

export async function enterDocPage({ params }: DocPageProps): Promise<string> {
	const { locale } = await params;
	setRequestLocale(locale);
	return locale;
}
