import type { Metadata } from "next";
import { redirect } from "next/navigation";
import { getTranslations, setRequestLocale } from "next-intl/server";
import { AuthCard } from "@/components/auth/auth-card";
import { getPathname, Link } from "@/i18n/navigation";
import { features } from "@/lib/env";
import { safeNextPath } from "@/lib/safe-next";
import { getSessionUser } from "@/lib/session";
import { LoginForm } from "./login-form";

type Props = { params: Promise<{ locale: string }>; searchParams: Promise<{ next?: string }> };

export async function generateMetadata({ params }: Props): Promise<Metadata> {
	const { locale } = await params;
	const t = await getTranslations({ locale, namespace: "auth.login" });
	return { title: t("metaTitle"), robots: { index: false } };
}

export default async function LoginPage({ params, searchParams }: Props) {
	const { locale } = await params;
	setRequestLocale(locale);
	const next = safeNextPath((await searchParams).next, getPathname({ href: "/", locale }));
	// `next` is already a localized path: the plain Next redirect avoids a second locale prefix.
	if (await getSessionUser()) redirect(next);
	const t = await getTranslations("auth.login");

	return (
		<AuthCard
			title={t("title")}
			footer={
				<>
					{t("noAccount")}{" "}
					<Link href="/signup" className="font-medium text-fg underline underline-offset-4">
						{t("createAccount")}
					</Link>
				</>
			}
		>
			<LoginForm next={next} googleEnabled={features.google} />
		</AuthCard>
	);
}
