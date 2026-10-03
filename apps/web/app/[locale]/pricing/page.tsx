import type { Metadata } from "next";
import { getTranslations, setRequestLocale } from "next-intl/server";
import { Link } from "@/i18n/navigation";
import { CREDIT_PACK_REQUESTS, PLAN_IDS, PLAN_LIMITS, type PlanId } from "@/lib/api/plans";
import {
	CREDIT_PACK_PRICE_CENTS,
	CURRENCIES,
	type Currency,
	currencyForLocale,
	currencySchema,
	formatPrice,
	MONTHLY_PRICE_CENTS,
} from "@/lib/billing/catalog";
import { features } from "@/lib/env";
import { getSessionUser } from "@/lib/session";

type Props = {
	params: Promise<{ locale: string }>;
	searchParams: Promise<Record<string, string | string[] | undefined>>;
};

export async function generateMetadata({ params }: Props): Promise<Metadata> {
	const { locale } = await params;
	const t = await getTranslations({ locale, namespace: "billing.pricing" });
	return { title: t("metaTitle") };
}

function priceOf(plan: PlanId, currency: Currency): number {
	return plan === "free" ? 0 : MONTHLY_PRICE_CENTS[plan][currency];
}

export default async function PricingPage({ params, searchParams }: Props) {
	const { locale } = await params;
	setRequestLocale(locale);
	const requested = currencySchema.safeParse((await searchParams).currency);
	const currency = requested.success ? requested.data : currencyForLocale(locale);
	const [t, tp, user] = await Promise.all([
		getTranslations("billing.pricing"),
		getTranslations("billing.plans"),
		getSessionUser(),
	]);
	const number = new Intl.NumberFormat(locale);
	const payments = features.stripe || features.pix;

	return (
		<section className="space-y-8">
			<header className="max-w-2xl space-y-3">
				<h1 className="font-serif text-3xl font-bold sm:text-4xl">{t("title")}</h1>
				<p className="text-muted">{t("subtitle")}</p>
			</header>

			<nav aria-label={t("currency")} className="flex flex-wrap items-center gap-2 text-sm">
				<span className="text-muted">{t("currency")}:</span>
				{CURRENCIES.map((c) => (
					<Link
						key={c}
						href={{ pathname: "/pricing", query: { currency: c } }}
						aria-current={c === currency ? "true" : undefined}
						className="inline-flex min-h-11 min-w-11 items-center justify-center rounded-md border border-border px-3 font-mono aria-[current=true]:bg-fg aria-[current=true]:text-bg"
					>
						{c.toUpperCase()}
					</Link>
				))}
			</nav>

			{!payments && (
				<p role="status" className="rounded-md border border-border bg-surface p-4 text-sm">
					{t("notConfigured")}
				</p>
			)}

			<ul className="grid gap-4 md:grid-cols-3">
				{PLAN_IDS.map((plan) => {
					const limits = PLAN_LIMITS[plan];
					const paid = plan !== "free";
					return (
						<li key={plan} className="flex flex-col rounded-lg border border-border p-5">
							<h2 className="font-serif text-2xl font-bold">{tp(plan)}</h2>
							<p className="mt-2">
								<span className="font-mono text-3xl tabular-nums">
									{formatPrice(priceOf(plan, currency), currency, locale)}
								</span>
								{paid && <span className="text-muted">{t("perMonth")}</span>}
							</p>
							<dl className="mt-4 space-y-2 text-sm">
								<dt className="sr-only">Quota</dt>
								<dd>
									{limits.period === "lifetime"
										? t("requestsLifetime", { count: number.format(limits.quota) })
										: t("requestsMonthly", { count: number.format(limits.quota) })}
								</dd>
								<dt className="sr-only">Page size</dt>
								<dd>{t("maxItems", { count: limits.maxPageSize })}</dd>
								<dt className="sr-only">Rate limit</dt>
								<dd>{t("rate", { count: limits.requestsPerSecond })}</dd>
								{paid && (
									<>
										<dt className="sr-only">Credits</dt>
										<dd>{t("creditsOnly")}</dd>
									</>
								)}
							</dl>
							<div className="mt-auto pt-6">
								{!paid ? (
									<Link
										href={user ? "/profile/api" : "/signup"}
										className="inline-flex min-h-11 w-full items-center justify-center rounded-md border border-border px-4"
									>
										{user ? t("ctaFreeLogged") : t("ctaFree")}
									</Link>
								) : (
									<Link
										href={
											user
												? { pathname: "/profile/api/checkout", query: { plan, currency } }
												: "/signup"
										}
										aria-disabled={!payments || undefined}
										className="inline-flex min-h-11 w-full items-center justify-center rounded-md bg-fg px-4 text-bg aria-disabled:pointer-events-none aria-disabled:opacity-50"
									>
										{user ? t("ctaPaid", { plan: tp(plan) }) : t("ctaSignup")}
									</Link>
								)}
							</div>
						</li>
					);
				})}
			</ul>

			<aside className="grid gap-4 md:grid-cols-2">
				<div className="rounded-lg border border-border p-5">
					<h2 className="font-serif text-xl font-bold">{t("creditsTitle")}</h2>
					<p className="mt-1 font-mono text-2xl tabular-nums">
						{formatPrice(CREDIT_PACK_PRICE_CENTS[currency], currency, locale)}
					</p>
					<p className="mt-2 text-sm text-muted">
						{t("creditsBody", { count: number.format(CREDIT_PACK_REQUESTS) })}
					</p>
				</div>
				<p className="rounded-lg border border-border p-5 text-sm">{t("pixNote")}</p>
			</aside>
		</section>
	);
}
