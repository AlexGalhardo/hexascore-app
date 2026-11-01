import { Link, usePage } from "@inertiajs/react";

import CoverageTable, { type Coverage } from "../../components/docs/CoverageTable";
import Seo from "../../components/Seo";
import { btnClass } from "../../components/ui/Btn";
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
	coverage: Coverage & { docsPath: string };
}

const eyebrow = "text-xs font-bold tracking-[0.1em] uppercase";
// The middle plan carries the "most chosen" mark of the frame.
const FEATURED: Plan["id"] = "pro";

export default function Index({ plans, currency, currencies, coverage }: Props) {
	const t = useT();
	const { locale, auth } = usePage<SharedProps>().props;
	const user = auth?.user ?? null;
	const number = new Intl.NumberFormat(locale);
	const money = (cents: number) =>
		new Intl.NumberFormat(locale, {
			style: "currency",
			currency: currency.toUpperCase(),
			minimumFractionDigits: cents % 100 === 0 ? 0 : 2,
		}).format(cents / 100);

	return (
		<section>
			<Seo title={t("billing.pricing.metaTitle")} />
			<div className="flex flex-wrap items-end justify-between gap-6">
				<header className="max-w-3xl">
					<p className={`${eyebrow} text-primary`}>{t("billing.pricing.eyebrow")}</p>
					<h1 className="mt-3 text-5xl leading-none md:text-7xl">
						{t("billing.pricing.title")}
					</h1>
					<p className="mt-3 text-lg text-muted">{t("billing.pricing.subtitle")}</p>
				</header>
				<nav aria-labelledby="currency-label">
					<p id="currency-label" className="mb-2 text-xs font-semibold">
						{t("billing.pricing.currency")}
					</p>
					<ul className="flex">
						{currencies.map((c) => (
							<li key={c}>
								<Link
									href={`${localized("/pricing", locale)}?currency=${c}`}
									aria-current={c === currency ? "true" : undefined}
									className="inline-flex min-h-11 min-w-22 items-center justify-center px-4 font-semibold hover:bg-hover focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fg aria-[current=true]:bg-fg aria-[current=true]:text-bg"
								>
									{c.toUpperCase()}
								</Link>
							</li>
						))}
					</ul>
				</nav>
			</div>

			<ul className="mt-10 grid gap-6 md:grid-cols-3">
				{plans.map((plan) => {
					const paid = plan.id !== "free";
					const featured = plan.id === FEATURED;
					const name = t(`billing.plans.${plan.id}`);
					return (
						<li
							key={plan.id}
							aria-labelledby={`plan-${plan.id}`}
							className={`flex flex-col border-t-2 border-fg p-6 ${featured ? "bg-surface" : ""}`}
						>
							<h2
								id={`plan-${plan.id}`}
								className="flex flex-wrap items-center gap-3 text-3xl"
							>
								{name}
								{featured && (
									<span
										className={`${eyebrow} bg-fg px-2 py-1 font-sans text-bg`}
									>
										{t("billing.pricing.popular")}
									</span>
								)}
							</h2>
							<p className="mt-6 flex items-baseline gap-2">
								<span className="font-display text-6xl leading-none tabular-nums">
									{money(paid ? (plan.amounts[currency] ?? 0) : 0)}
								</span>
								<span className="text-muted">
									{paid
										? t("billing.pricing.perMonth")
										: t("billing.pricing.lifetime")}
								</span>
							</p>
							<dl className="mt-6">
								{[
									[
										"billing.pricing.quotaLabel",
										t(
											plan.lifetime
												? "billing.pricing.requestsLifetime"
												: "billing.pricing.requestsMonthly",
											{ count: number.format(plan.quota) },
										),
									],
									[
										"billing.pricing.pageSizeLabel",
										t("billing.pricing.maxItems", { count: plan.maxPageSize }),
									],
									[
										"billing.pricing.rateLabel",
										t("billing.pricing.rate", { count: plan.rps }),
									],
								].map(([label, value]) => (
									<div key={label} className="border-t border-border py-3">
										<dt className={`${eyebrow} text-muted`}>{t(label)}</dt>
										<dd className="mt-1">{value}</dd>
									</div>
								))}
							</dl>
							<div className="mt-auto pt-4">
								{paid ? (
									<Link
										href={
											user
												? localized("/billing", locale)
												: localized("/signup", locale)
										}
										className={btnClass(
											featured ? "primary" : "secondary",
											"w-full",
										)}
									>
										{user
											? t("billing.pricing.ctaPaid", { plan: name })
											: t("billing.pricing.ctaSignup")}
									</Link>
								) : (
									<Link
										href={
											user
												? localized("/profile/api", locale)
												: localized("/signup", locale)
										}
										className={btnClass("secondary", "w-full")}
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

			<aside aria-labelledby="coverage-h" className="mt-10 border-t-2 border-fg pt-4">
				<h2 id="coverage-h" className="text-3xl">
					{t("billing.pricing.coverageTitle")}
				</h2>
				<p className="mt-2">{t("billing.pricing.coverageBody")}</p>
				<CoverageTable coverage={coverage} />
				<a
					href={localized(coverage.docsPath, locale)}
					className="mt-3 inline-flex min-h-11 items-center gap-1 font-semibold text-primary hover:underline"
				>
					{t("billing.pricing.coverageLink")} <span aria-hidden="true">→</span>
				</a>
			</aside>
		</section>
	);
}
