import type { Metadata } from "next";
import { getTranslations, setRequestLocale } from "next-intl/server";
import { getSessionUser } from "@/lib/session";
import { ContactForm } from "./contact-form";

type Props = { params: Promise<{ locale: string }> };

export async function generateMetadata({ params }: Props): Promise<Metadata> {
	const { locale } = await params;
	const t = await getTranslations({ locale, namespace: "contact" });
	return { title: t("metaTitle") };
}

export default async function ContactPage({ params }: Props) {
	const { locale } = await params;
	setRequestLocale(locale);
	const [me, t] = await Promise.all([getSessionUser(), getTranslations("contact")]);

	return (
		<section className="mx-auto flex max-w-xl flex-col gap-6">
			<div>
				<h1 className="font-serif text-3xl font-bold">{t("title")}</h1>
				<p className="mt-2 text-muted">{t("intro")}</p>
			</div>
			<ContactForm user={me ? { name: me.name, email: me.email } : null} />
		</section>
	);
}
