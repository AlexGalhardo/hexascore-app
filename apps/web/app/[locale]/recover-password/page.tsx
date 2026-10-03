import type { Metadata } from "next";
import { getTranslations, setRequestLocale } from "next-intl/server";
import { AuthCard } from "@/components/auth/auth-card";
import { EmailRequestForm } from "@/components/auth/email-request-form";
import { Link } from "@/i18n/navigation";

type Props = { params: Promise<{ locale: string }> };

export async function generateMetadata({ params }: Props): Promise<Metadata> {
	const { locale } = await params;
	const t = await getTranslations({ locale, namespace: "auth.recover" });
	return { title: t("metaTitle") };
}

export default async function RecoverPasswordPage({ params }: Props) {
	const { locale } = await params;
	setRequestLocale(locale);
	const t = await getTranslations("auth.recover");

	return (
		<AuthCard
			title={t("title")}
			footer={
				<Link href="/login" className="font-medium text-fg underline underline-offset-4">
					{t("back")}
				</Link>
			}
		>
			<p className="text-muted">{t("body")}</p>
			<EmailRequestForm kind="recover" />
		</AuthCard>
	);
}
