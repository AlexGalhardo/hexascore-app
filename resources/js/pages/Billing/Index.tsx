import { Head, Link, usePage } from "@inertiajs/react";
import { useState } from "react";

import { useT } from "../../lib/useT";

type Currency = "brl" | "usd" | "eur";

interface Plan {
	name: string;
	label: string;
	current: boolean;
	amounts: Record<Currency, number>;
}

interface Subscription {
	plan: string;
	status: string;
	state: string;
	valid: boolean;
	currentPeriodEnd: string;
	cancelAtPeriodEnd: boolean;
}

interface Props {
	plans: Plan[];
	subscription: Subscription | null;
	subscribed: boolean;
	checkout: boolean | "canceled";
	currency: Currency;
}

const currencies: Currency[] = ["usd", "brl", "eur"];
const button =
	"rounded-md bg-primary px-4 py-2 font-medium text-on-primary hover:bg-primary-hover disabled:opacity-50";

// Checkout and the portal are Stripe pages: the server answers these
// visits with a full page redirect (inertia.Location).
export default function Billing({ plans, subscription, subscribed, checkout, currency }: Props) {
	const t = useT();
	const { locale } = usePage<{ locale: string }>().props;
	const [selected, setSelected] = useState<Currency>(currency);
	const money = (minor: number) =>
		new Intl.NumberFormat(locale, {
			style: "currency",
			currency: selected.toUpperCase(),
		}).format(minor / 100);

	return (
		<>
			<Head title={t("billing.checkout.metaTitle")} />
			<h1 className="mb-4 text-2xl font-semibold">{t("billing.checkout.title")}</h1>
			{checkout === true && <p className="mb-4">{t("billing.checkout.success")}</p>}
			{checkout === "canceled" && <p className="mb-4">{t("billing.checkout.canceled")}</p>}
			{subscription && (
				<div className="space-y-4">
					<p>
						{t("billing.checkout.current")}: <strong>{subscription.plan}</strong> (
						{t(`billing.checkout.status.${subscription.state}`)})
						{subscription.currentPeriodEnd &&
							` - ${t(
								subscription.cancelAtPeriodEnd
									? "billing.checkout.ends"
									: "billing.checkout.renews",
								{
									date: subscription.currentPeriodEnd,
								},
							)}`}
					</p>
					<Link href="/billing/portal" method="post" as="button" className={button}>
						{t("billing.checkout.manage")}
					</Link>
				</div>
			)}
			{!subscribed && (
				<>
					<label className="mt-6 flex items-center gap-2">
						{t("billing.checkout.currency")}
						<select
							value={selected}
							onChange={(e) => setSelected(e.target.value as Currency)}
							className="rounded-md border px-2 py-1"
						>
							{currencies.map((c) => (
								<option key={c} value={c}>
									{c.toUpperCase()}
								</option>
							))}
						</select>
					</label>
					<ul className="space-y-2 pt-4">
						{plans.map((plan) => (
							<li key={plan.name}>
								<Link
									href={`/billing/checkout/${plan.name}`}
									method="post"
									data={{ currency: selected }}
									as="button"
									className={button}
								>
									{t("billing.checkout.buy")} {t(`billing.plans.${plan.name}`)} -{" "}
									{money(plan.amounts[selected])}
									{t("billing.pricing.perMonth")}
								</Link>
							</li>
						))}
					</ul>
				</>
			)}
		</>
	);
}
