import type { Metadata, Route } from "next";
import { redirect } from "next/navigation";
import { getTranslations, setRequestLocale } from "next-intl/server";
import { AuthCard } from "@/components/auth/auth-card";
import { getPathname, redirect as localizedRedirect } from "@/i18n/navigation";
import { safeNextPath } from "@/lib/safe-next";
import { getSessionUser } from "@/lib/session";
import { NicknameForm } from "./nickname-form";

type Props = { params: Promise<{ locale: string }>; searchParams: Promise<{ next?: string | string[] }> };

export async function generateMetadata({ params }: Props): Promise<Metadata> {
	const { locale } = await params;
	const t = await getTranslations({ locale, namespace: "social.onboarding" });
	return { title: t("title"), robots: { index: false } };
}

// Google sign-ups arrive without a nickname; pages that need one (rating, feed, profile) send them here.
export default async function OnboardingPage({ params, searchParams }: Props) {
	const { locale } = await params;
	setRequestLocale(locale);
	const raw = (await searchParams).next;
	const next = safeNextPath(typeof raw === "string" ? raw : null, getPathname({ href: "/profile", locale }));
	const me = await getSessionUser();
	if (!me) {
		return localizedRedirect({
			href: { pathname: "/login", query: { next: getPathname({ href: "/onboarding", locale }) } },
			locale,
		});
	}
	// `next` is already localized, so next/navigation's redirect (no second locale prefix).
	if (me.username) redirect(next as Route);
	const t = await getTranslations("social.onboarding");

	return (
		<AuthCard title={t("title")}>
			<p className="text-muted">{t("body")}</p>
			<NicknameForm next={next} />
		</AuthCard>
	);
}
