import Stripe from "stripe";
import { env } from "../env";
import { type PaidPlan, planFromLookupKey, STRIPE_LOOKUP_KEYS, type StripeProduct } from "./catalog";

export const PORTAL_CONFIG_METADATA = { app: "hexascore" } as const;

let client: Stripe | null | undefined;

/** null when Stripe keys are not configured: callers show "payments not configured" instead of crashing. */
export function getStripe(): Stripe | null {
	if (client === undefined) {
		client = env.STRIPE_SECRET_KEY
			? new Stripe(env.STRIPE_SECRET_KEY, { maxNetworkRetries: 2, timeout: 10_000 })
			: null;
	}
	return client;
}

// Prices are resolved by lookup_key (created by scripts/stripe-setup.ts), so no price ids live in env.
const priceCache = new Map<StripeProduct, { id: string; at: number }>();
const PRICE_TTL_MS = 10 * 60_000;

export async function priceIdFor(stripe: Stripe, product: StripeProduct): Promise<string | null> {
	const cached = priceCache.get(product);
	if (cached && Date.now() - cached.at < PRICE_TTL_MS) return cached.id;
	const prices = await stripe.prices.list({
		lookup_keys: Object.values(STRIPE_LOOKUP_KEYS),
		active: true,
		limit: 10,
	});
	for (const price of prices.data) {
		const entry = Object.entries(STRIPE_LOOKUP_KEYS).find(([, key]) => key === price.lookup_key);
		if (entry) priceCache.set(entry[0] as StripeProduct, { id: price.id, at: Date.now() });
	}
	return priceCache.get(product)?.id ?? null;
}

let portalConfigId: string | null | undefined;

/** The portal configuration created by stripe-setup (plan switching); falls back to the account default. */
export async function portalConfigurationId(stripe: Stripe): Promise<string | undefined> {
	if (portalConfigId === undefined) {
		const configs = await stripe.billingPortal.configurations.list({ active: true, limit: 100 });
		portalConfigId = configs.data.find((c) => c.metadata?.app === PORTAL_CONFIG_METADATA.app)?.id ?? null;
	}
	return portalConfigId ?? undefined;
}

export type LocalSubscriptionStatus = "active" | "past_due" | "canceled";

export function mapStripeStatus(status: Stripe.Subscription.Status): LocalSubscriptionStatus {
	switch (status) {
		case "active":
		case "trialing":
			return "active";
		case "canceled":
		case "incomplete_expired":
			return "canceled";
		default:
			// past_due, unpaid, incomplete, paused: keep the row but do not grant the plan.
			return "past_due";
	}
}

export type NormalizedSubscription = {
	subscriptionId: string;
	customerId: string;
	metadataUserId: string | null;
	plan: PaidPlan | null;
	status: LocalSubscriptionStatus;
	currentPeriodEnd: Date;
	cancelAtPeriodEnd: boolean;
};

/** Since API 2025-03-31 the billing period lives on subscription items, not on the subscription. */
export function normalizeSubscription(sub: Stripe.Subscription): NormalizedSubscription {
	const items = sub.items.data;
	const periodEnd = items.length ? Math.max(...items.map((i) => i.current_period_end)) : (sub.ended_at ?? 0);
	const plan = items.map((i) => planFromLookupKey(i.price.lookup_key)).find((p) => p !== null) ?? null;
	const userId = sub.metadata?.user_id;
	return {
		subscriptionId: sub.id,
		customerId: typeof sub.customer === "string" ? sub.customer : sub.customer.id,
		metadataUserId: typeof userId === "string" && userId ? userId : null,
		plan,
		status: mapStripeStatus(sub.status),
		currentPeriodEnd: new Date(periodEnd * 1000),
		cancelAtPeriodEnd: sub.cancel_at_period_end || sub.cancel_at !== null,
	};
}
