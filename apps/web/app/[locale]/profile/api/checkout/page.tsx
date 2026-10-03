import type { Metadata } from "next";
import { getTranslations, setRequestLocale } from "next-intl/server";
import { CheckoutForm } from "@/components/billing/checkout-form";
import { ManageBillingButton } from "@/components/billing/manage-billing-button";
import { redirect } from "@/i18n/navigation";
import { currencyForLocale, currencySchema } from "@/lib/billing/catalog";
import { getBillingSummary } from "@/lib/billing/summary";
import { features } from "@/lib/env";
import { pixCheckoutFlag } from "@/lib/flags";
import { getSessionUser } from "@/lib/session";

type Props = {
	params: Promise<{ locale: string }>;
	searchParams: Promise<Record<string, string | string[] | undefined>>;
};

export async function generateMetadata({ params }: Props): Promise<Metadata> {
	const { locale } = await params;
	const t = await getTranslations({ locale, namespace: "billing.checkout" });
	return { title: t("metaTitle"), robots: { index: false } };
}

export default async function CheckoutPage({ params, searchParams }: Props) {
	const { locale } = await params;
	setRequestLocale(locale);
	const user = await getSessionUser();
	if (!user) {
		redirect({ href: { pathname: "/login", query: { next: "/profile/api/checkout" } }, locale });
		return null;
	}
	const query = await searchParams;
	const requested = currencySchema.safeParse(query.currency);
	const currency = requested.success ? requested.data : currencyForLocale(locale);
	const product = query.plan === "business" || query.plan === "credits" ? query.plan : "pro";
	const status = query.status === "success" || query.status === "canceled" ? query.status : null;

	const [t, tp, summary, pixFlag] = await Promise.all([
		getTranslations("billing.checkout"),
		getTranslations("billing.plans"),
		getBillingSummary(user.id),
		pixCheckoutFlag(),
	]);
	const pix = features.pix && pixFlag;
	const date = (iso: string) => new Intl.DateTimeFormat(locale, { dateStyle: "long" }).format(new Date(iso));
	const renews = summary.provider === "stripe" && !summary.cancelAtPeriodEnd;

	return (
		<section className="max-w-3xl space-y-8">
			<h1 className="font-serif text-3xl font-bold sm:text-4xl">{t("title")}</h1>

			{status && (
				<p role="status" className="rounded-md border border-border bg-surface p-4 text-sm">
					{t(status)}
				</p>
			)}

			<section aria-labelledby="current-plan" className="rounded-lg border border-border p-5">
				<h2 id="current-plan" className="text-sm text-muted">
					{t("current")}
				</h2>
				<p className="mt-1 font-serif text-2xl font-bold">
					{tp(summary.plan)}
					{summary.status && summary.plan !== "free" && summary.provider && (
						<span className="ml-2 align-middle font-sans text-sm font-normal text-muted">
							{t(`status.${summary.status}`)} · {t(`via.${summary.provider}`)}
						</span>
					)}
				</p>
				{summary.plan !== "free" && summary.currentPeriodEnd && (
					<p className="mt-1 text-sm">
						{renews
							? t("renews", { date: date(summary.currentPeriodEnd) })
							: t("ends", { date: date(summary.currentPeriodEnd) })}
					</p>
				)}
				<p className="mt-1 font-mono text-sm tabular-nums">
					{t("credits", { count: new Intl.NumberFormat(locale).format(summary.credits) })}
				</p>
				{features.stripe && summary.hasCardBilling && (
					<div className="mt-4">
						<ManageBillingButton />
					</div>
				)}
			</section>

			{features.stripe || pix ? (
				<CheckoutForm
					stripe={features.stripe}
					pix={pix}
					defaultCurrency={currency}
					defaultProduct={product}
					hasPaidPlan={summary.plan !== "free"}
				/>
			) : (
				<p role="status" className="rounded-md border border-border bg-surface p-4 text-sm">
					{t("notConfigured")}
				</p>
			)}
		</section>
	);
}
