import { z } from "zod";
import { CREDIT_PACK_REQUESTS } from "../api/plans";

// Single source of every price the server charges (minor units). Clients only ever send ids, never amounts.

export const CURRENCIES = ["brl", "usd", "eur"] as const;
export type Currency = (typeof CURRENCIES)[number];
export const currencySchema = z.enum(CURRENCIES);

export const PAID_PLANS = ["pro", "business"] as const;
export type PaidPlan = (typeof PAID_PLANS)[number];
export const paidPlanSchema = z.enum(PAID_PLANS);

export const MONTHLY_PRICE_CENTS: Record<PaidPlan, Record<Currency, number>> = {
	pro: { brl: 2900, usd: 600, eur: 600 },
	business: { brl: 14900, usd: 2900, eur: 2700 },
};

export const CREDIT_PACK_PRICE_CENTS: Record<Currency, number> = { brl: 1900, usd: 400, eur: 400 };

/** PIX annual: 12 months for the price of this many. */
export const PIX_ANNUAL_CHARGED_MONTHS = 10;

export const PIX_PURPOSES = ["pro_1m", "pro_12m", "business_1m", "business_12m", "credits"] as const;
export type PixPurpose = (typeof PIX_PURPOSES)[number];
export const pixPurposeSchema = z.enum(PIX_PURPOSES);

export type PixOffer =
	| { kind: "plan"; plan: PaidPlan; months: 1 | 12; amountCents: number }
	| { kind: "credits"; requests: number; amountCents: number };

export function pixOffer(purpose: PixPurpose): PixOffer {
	if (purpose === "credits") {
		return { kind: "credits", requests: CREDIT_PACK_REQUESTS, amountCents: CREDIT_PACK_PRICE_CENTS.brl };
	}
	const [plan, period] = purpose.split("_") as [PaidPlan, "1m" | "12m"];
	const monthly = MONTHLY_PRICE_CENTS[plan].brl;
	return period === "12m"
		? { kind: "plan", plan, months: 12, amountCents: monthly * PIX_ANNUAL_CHARGED_MONTHS }
		: { kind: "plan", plan, months: 1, amountCents: monthly };
}

export function currencyForLocale(locale: string): Currency {
	if (locale === "pt-BR") return "brl";
	if (locale === "es") return "eur";
	return "usd";
}

export const STRIPE_LOOKUP_KEYS = {
	pro: "hexascore_pro_monthly",
	business: "hexascore_business_monthly",
	credits: "hexascore_credits_25k",
} as const;
export type StripeProduct = keyof typeof STRIPE_LOOKUP_KEYS;

export function planFromLookupKey(key: string | null | undefined): PaidPlan | null {
	if (key === STRIPE_LOOKUP_KEYS.pro) return "pro";
	if (key === STRIPE_LOOKUP_KEYS.business) return "business";
	return null;
}

export function formatPrice(cents: number, currency: Currency, locale: string): string {
	return new Intl.NumberFormat(locale, { style: "currency", currency: currency.toUpperCase() }).format(cents / 100);
}

/** Adds calendar months in UTC, clamping to the last day (Jan 31 + 1 month = Feb 28/29). */
export function addMonths(date: Date, months: number): Date {
	const d = new Date(date);
	const day = d.getUTCDate();
	d.setUTCDate(1);
	d.setUTCMonth(d.getUTCMonth() + months);
	const lastDay = new Date(Date.UTC(d.getUTCFullYear(), d.getUTCMonth() + 1, 0)).getUTCDate();
	d.setUTCDate(Math.min(day, lastDay));
	return d;
}

/** PIX renewals stack: paying early never loses the days already paid for. */
export function extendPeriod(currentEnd: Date | null, now: Date, months: number): Date {
	const base = currentEnd && currentEnd.getTime() > now.getTime() ? currentEnd : now;
	return addMonths(base, months);
}
