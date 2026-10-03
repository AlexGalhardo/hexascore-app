import type { Metadata } from "next";
import { getTranslations, setRequestLocale } from "next-intl/server";
import { AuthCard } from "@/components/auth/auth-card";
import { EmailRequestForm } from "@/components/auth/email-request-form";
import { buttonClass, Notice } from "@/components/ui/form";
import { Link } from "@/i18n/navigation";
import { getSessionUser } from "@/lib/session";

type Props = { params: Promise<{ locale: string }>; searchParams: Promise<{ error?: string }> };

export async function generateMetadata({ params }: Props): Promise<Metadata> {
	const { locale } = await params;
	const t = await getTranslations({ locale, namespace: "auth.verify" });
	return { title: t("metaTitle"), robots: { index: false } };
}

// Landing page of the confirmation link: Better Auth verifies, signs in and redirects here (or adds ?error=).
export default async function VerifyEmailPage({ params, searchParams }: Props) {
	const { locale } = await params;
	setRequestLocale(locale);
	const { error } = await searchParams;
	const user = await getSessionUser();
	const t = await getTranslations("auth.verify");

	if (!error && user?.emailVerified) {
		return (
			<AuthCard title={t("success")}>
				<Notice tone="success">{t("successBody")}</Notice>
				<Link href="/profile" className={buttonClass("primary")}>
					{t("continue")}
				</Link>
			</AuthCard>
		);
	}

	return (
		<AuthCard title={t("title")}>
			{error ? <Notice tone="error">{t("invalid")}</Notice> : <p className="text-muted">{t("pending")}</p>}
			<EmailRequestForm kind="verify" defaultEmail={user?.email} />
		</AuthCard>
	);
}
