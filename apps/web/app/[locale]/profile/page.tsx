import { and, asc, eq, isNotNull } from "drizzle-orm";
import type { Metadata } from "next";
import { getTranslations, setRequestLocale } from "next-intl/server";
import type { ReactNode } from "react";
import { getPathname, redirect } from "@/i18n/navigation";
import { db } from "@/lib/db/client";
import { account, passkey } from "@/lib/db/schema";
import { getSessionUser } from "@/lib/session";
import { DeleteAccount } from "./delete-account";
import { DetailsForm } from "./details-form";
import { PasskeysSection } from "./passkeys-section";
import { PasswordSection } from "./password-section";
import { PrivacyToggle } from "./privacy-toggle";
import { TwoFactorSection } from "./two-factor-section";

type Props = { params: Promise<{ locale: string }> };

export async function generateMetadata({ params }: Props): Promise<Metadata> {
	const { locale } = await params;
	const t = await getTranslations({ locale, namespace: "account" });
	return { title: t("metaTitle"), robots: { index: false } };
}

function Section({ title, children }: { title: string; children: ReactNode }) {
	return (
		<section className="flex flex-col gap-4 border-t border-border pt-6">
			<h2 className="font-serif text-xl font-bold">{title}</h2>
			{children}
		</section>
	);
}

export default async function ProfilePage({ params }: Props) {
	const { locale } = await params;
	setRequestLocale(locale);
	const me = await getSessionUser();
	if (!me) {
		return redirect({
			href: { pathname: "/login", query: { next: getPathname({ href: "/profile", locale }) } },
			locale,
		});
	}
	if (!me.username) {
		return redirect({
			href: { pathname: "/onboarding", query: { next: getPathname({ href: "/profile", locale }) } },
			locale,
		});
	}

	const [credential, keys] = await Promise.all([
		db
			.select({ id: account.id })
			.from(account)
			.where(and(eq(account.userId, me.id), eq(account.providerId, "credential"), isNotNull(account.password)))
			.limit(1),
		db
			.select({ id: passkey.id, name: passkey.name, createdAt: passkey.createdAt })
			.from(passkey)
			.where(eq(passkey.userId, me.id))
			.orderBy(asc(passkey.createdAt)),
	]);
	const hasPassword = credential.length > 0;
	const t = await getTranslations("account");

	return (
		<div className="mx-auto flex max-w-2xl flex-col gap-8">
			<h1 className="font-serif text-3xl font-bold">{t("title")}</h1>
			<Section title={t("details.title")}>
				<DetailsForm name={me.name} email={me.email} nickname={me.username} />
			</Section>
			<Section title={hasPassword ? t("password.changeTitle") : t("password.setTitle")}>
				<PasswordSection hasPassword={hasPassword} />
			</Section>
			<Section title={t("twoFactor.title")}>
				<TwoFactorSection enabled={me.twoFactorEnabled} needsPassword={hasPassword} />
			</Section>
			<Section title={t("passkeys.title")}>
				<PasskeysSection
					passkeys={keys.map((k) => ({
						id: k.id,
						name: k.name,
						createdAt: k.createdAt?.toISOString() ?? null,
					}))}
				/>
			</Section>
			<Section title={t("privacy.title")}>
				<PrivacyToggle initial={me.profilePrivate} />
			</Section>
			<Section title={t("delete.title")}>
				<DeleteAccount />
			</Section>
		</div>
	);
}
