import type { Metadata } from "next";
import { getTranslations, setRequestLocale } from "next-intl/server";
import { AuthCard } from "@/components/auth/auth-card";
import { buttonClass, Notice } from "@/components/ui/form";
import { Link } from "@/i18n/navigation";
import { ResetPasswordForm } from "./reset-password-form";

type Props = { params: Promise<{ locale: string }>; searchParams: Promise<{ token?: string; error?: string }> };

export async function generateMetadata({ params }: Props): Promise<Metadata> {
	const { locale } = await params;
	const t = await getTranslations({ locale, namespace: "auth.reset" });
	return { title: t("metaTitle"), robots: { index: false } };
}

// Better Auth validates the emailed link and redirects here with ?token= (or ?error=INVALID_TOKEN).
export default async function ResetPasswordPage({ params, searchParams }: Props) {
	const { locale } = await params;
	setRequestLocale(locale);
	const { token, error } = await searchParams;
	const t = await getTranslations("auth.reset");

	return (
		<AuthCard title={t("title")}>
			{error || !token ? (
				<>
					<Notice tone="error">{t("invalid")}</Notice>
					<Link href="/recover-password" className={buttonClass("secondary")}>
						{t("requestNew")}
					</Link>
				</>
			) : (
				<ResetPasswordForm token={token} />
			)}
		</AuthCard>
	);
}
