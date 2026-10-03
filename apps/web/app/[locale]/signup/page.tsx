import type { Metadata } from "next";
import { getTranslations, setRequestLocale } from "next-intl/server";
import { AuthCard } from "@/components/auth/auth-card";
import { Link, redirect } from "@/i18n/navigation";
import { features } from "@/lib/env";
import { getSessionUser } from "@/lib/session";
import { SignupForm } from "./signup-form";

type Props = { params: Promise<{ locale: string }> };

export async function generateMetadata({ params }: Props): Promise<Metadata> {
	const { locale } = await params;
	const t = await getTranslations({ locale, namespace: "auth.signup" });
	return { title: t("metaTitle") };
}

export default async function SignupPage({ params }: Props) {
	const { locale } = await params;
	setRequestLocale(locale);
	if (await getSessionUser()) redirect({ href: "/profile", locale });
	const t = await getTranslations("auth.signup");

	return (
		<AuthCard
			title={t("title")}
			footer={
				<>
					{t("haveAccount")}{" "}
					<Link href="/login" className="font-medium text-fg underline underline-offset-4">
						{t("logIn")}
					</Link>
				</>
			}
		>
			<SignupForm googleEnabled={features.google} />
		</AuthCard>
	);
}
