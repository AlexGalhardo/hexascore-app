import { Link, usePage } from "@inertiajs/react";

import Seo from "../../components/Seo";
import { localized } from "../../lib/route";
import { useT } from "../../lib/useT";
import type { SharedProps } from "../../types";

interface Plan {
	id: "free" | "pro" | "business";
	amounts: Record<string, number>;
	quota: number;
	lifetime: boolean;
	maxPageSize: number;
	rps: number;
}

interface Props {
	plans: Plan[];
	currency: string;
	currencies: string[];
	coverage: { measured: boolean; docsPath: string };
}

const cta = "inline-flex min-h-11 w-full items-center justify-center rounded-md px-4 font-medium";

export default function Index({ plans, currency, currencies, coverage }: Props) {
	const t = useT();
	const { locale, auth } = usePage<SharedProps>().props;
	const user = auth?.user ?? null;
	const number = new Intl.NumberFormat(locale);
	const money = (cents: number) =>
		new Intl.NumberFormat(locale, {
			style: "currency",
			currency: currency.toUpperCase(),
		}).format(cents / 100);

	return (
		<section className="space-y-8">
			<Seo title={t("billing.pricing.metaTitle")} />
			<header className="max-w-2xl space-y-3">
				<h1 className="text-3xl font-bold">{t("billing.pricing.title")}</h1>
				<p>{t("billing.pricing.subtitle")}</p>
			</header>

			<nav
				aria-label={t("billing.pricing.currency")}
				className="flex flex-wrap items-center gap-2 text-sm"
			>
				<span>{t("billing.pricing.currency")}:</span>
				{currencies.map((c) => (
					<Link
						key={c}
						href={`${localized("/pricing", locale)}?currency=${c}`}
						aria-current={c === currency ? "true" : undefined}
						className="inline-flex min-h-11 min-w-11 items-center justify-center rounded-md border px-3 font-mono aria-[current=true]:bg-indigo-700 aria-[current=true]:text-white"
					>
						{c.toUpperCase()}
					</Link>
				))}
			</nav>

			<ul className="grid gap-4 md:grid-cols-3">
				{plans.map((plan) => {
					const paid = plan.id !== "free";
					return (
						<li key={plan.id} className="flex flex-col rounded-lg border p-5">
							<h2 className="text-2xl font-bold">{t(`billing.plans.${plan.id}`)}</h2>
							<p className="mt-2">
								<span className="font-mono text-3xl tabular-nums">
									{money(paid ? (plan.amounts[currency] ?? 0) : 0)}
								</span>
								{paid && <span>{t("billing.pricing.perMonth")}</span>}
							</p>
							<dl className="mt-4 space-y-2 text-sm">
								<dt className="sr-only">{t("billing.pricing.quotaLabel")}</dt>
								<dd>
									{t(
										plan.lifetime
											? "billing.pricing.requestsLifetime"
											: "billing.pricing.requestsMonthly",
										{ count: number.format(plan.quota) },
									)}
								</dd>
								<dt className="sr-only">{t("billing.pricing.pageSizeLabel")}</dt>
								<dd>
									{t("billing.pricing.maxItems", { count: plan.maxPageSize })}
								</dd>
								<dt className="sr-only">{t("billing.pricing.rateLabel")}</dt>
								<dd>{t("billing.pricing.rate", { count: plan.rps })}</dd>
							</dl>
							<div className="mt-auto pt-6">
								{paid ? (
									<Link
										href={
											user
												? localized("/billing", locale)
												: localized("/signup", locale)
										}
										className={`${cta} bg-indigo-700 text-white`}
									>
										{user
											? t("billing.pricing.ctaPaid", {
													plan: t(`billing.plans.${plan.id}`),
												})
											: t("billing.pricing.ctaSignup")}
									</Link>
								) : (
									<Link
										href={
											user
												? localized("/profile/api", locale)
												: localized("/signup", locale)
										}
										className={`${cta} border`}
									>
										{user
											? t("billing.pricing.ctaFreeLogged")
											: t("billing.pricing.ctaFree")}
									</Link>
								)}
							</div>
						</li>
					);
				})}
			</ul>

			<aside
				aria-labelledby="coverage-h"
				className="max-w-2xl space-y-2 rounded-lg border p-5"
			>
				<h2 id="coverage-h" className="text-xl font-bold">
					{t("billing.pricing.coverageTitle")}
				</h2>
				<p className="text-sm">{t("billing.pricing.coverageBody")}</p>
				<a href={localized(coverage.docsPath, locale)} className="text-sm underline">
					{t("billing.pricing.coverageLink")}
				</a>
			</aside>
		</section>
	);
}
