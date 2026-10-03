import { Analytics } from "@vercel/analytics/next";
import { SpeedInsights } from "@vercel/speed-insights/next";
import type { Metadata } from "next";
import { Inter, JetBrains_Mono, Source_Serif_4 } from "next/font/google";
import { headers } from "next/headers";
import { notFound } from "next/navigation";
import { hasLocale, NextIntlClientProvider } from "next-intl";
import { getTranslations, setRequestLocale } from "next-intl/server";
import type { ReactNode } from "react";
import { DeletionBanner } from "@/components/account/deletion-banner";
import { CookieBanner } from "@/components/layout/cookie-banner";
import { Footer } from "@/components/layout/footer";
import { Navbar } from "@/components/layout/navbar";
import { QueryProvider } from "@/components/layout/query-provider";
import { SandboxBanner } from "@/components/layout/sandbox-banner";
import { routing } from "@/i18n/routing";
import { APP_NAME, APP_URL } from "@/lib/config";
import { env } from "@/lib/env";
import { themeBootScript } from "@/lib/preferences";
import "../globals.css";

const display = Source_Serif_4({ subsets: ["latin"], variable: "--font-display" });
const text = Inter({ subsets: ["latin"], variable: "--font-text" });
const numeric = JetBrains_Mono({ subsets: ["latin"], variable: "--font-numeric" });

type Props = { children: ReactNode; params: Promise<{ locale: string }> };

export function generateStaticParams() {
	return routing.locales.map((locale) => ({ locale }));
}

export async function generateMetadata({ params }: Props): Promise<Metadata> {
	const { locale } = await params;
	const t = await getTranslations({ locale, namespace: "meta" });
	return {
		metadataBase: new URL(APP_URL),
		title: { default: APP_NAME, template: `%s · ${APP_NAME}` },
		description: t("description"),
		// Sandbox (fake users and ratings): never indexed, on top of robots.ts disallowing every path.
		...(env.NEXT_PUBLIC_SANDBOX && { robots: { index: false, follow: false } }),
	};
}

export default async function LocaleLayout({ children, params }: Props) {
	const { locale } = await params;
	if (!hasLocale(routing.locales, locale)) notFound();
	setRequestLocale(locale);
	const t = await getTranslations("cookies");
	// Set by proxy.ts. Reading it also makes every page dynamic, which a per-request nonce requires.
	const nonce = (await headers()).get("x-nonce") ?? undefined;

	return (
		<html
			lang={locale}
			suppressHydrationWarning
			className={`${display.variable} ${text.variable} ${numeric.variable}`}
		>
			<head>
				{/* biome-ignore lint/security/noDangerouslySetInnerHtml: static theme boot script, no user input */}
				<script nonce={nonce} dangerouslySetInnerHTML={{ __html: themeBootScript }} />
			</head>
			<body className="min-h-dvh font-sans antialiased">
				<NextIntlClientProvider>
					<QueryProvider>
						<SandboxBanner />
						<Navbar />
						<DeletionBanner />
						<main className="mx-auto max-w-6xl px-4 py-8">{children}</main>
						<Footer />
						<CookieBanner text={t("text")} accept={t("accept")} decline={t("decline")} />
					</QueryProvider>
				</NextIntlClientProvider>
				{/* Cookieless (no consent needed); scripts are injected by these nonced bundles, allowed by 'strict-dynamic'. */}
				<Analytics />
				<SpeedInsights />
			</body>
		</html>
	);
}
