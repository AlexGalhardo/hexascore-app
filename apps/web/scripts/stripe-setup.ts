/**
 * Idempotently creates the Stripe Products, Prices (one per plan/pack with BRL/USD/EUR currency_options) and the
 * Customer Portal configuration in whatever account STRIPE_SECRET_KEY points at. Run: `bun run stripe:setup`.
 * Prices are immutable: when an amount in lib/billing/catalog.ts changes, a new Price takes over the lookup_key
 * (transfer_lookup_key) and the old one is archived; existing subscribers keep their old Price until changed.
 */
import type Stripe from "stripe";
import {
	CREDIT_PACK_PRICE_CENTS,
	type Currency,
	MONTHLY_PRICE_CENTS,
	STRIPE_LOOKUP_KEYS,
	type StripeProduct,
} from "../lib/billing/catalog";
import { getStripe, PORTAL_CONFIG_METADATA } from "../lib/billing/stripe";
import { APP_NAME } from "../lib/config";

const NAMES: Record<StripeProduct, string> = {
	pro: `${APP_NAME} API Pro`,
	business: `${APP_NAME} API Business`,
	credits: `${APP_NAME} API credits (+25,000 requests)`,
};

function amounts(product: StripeProduct): Record<Currency, number> {
	return product === "credits" ? CREDIT_PACK_PRICE_CENTS : MONTHLY_PRICE_CENTS[product];
}

function matches(price: Stripe.Price, product: StripeProduct): boolean {
	const want = amounts(product);
	const recurring = product === "credits" ? price.recurring === null : price.recurring?.interval === "month";
	return (
		recurring &&
		price.currency === "brl" &&
		price.unit_amount === want.brl &&
		price.currency_options?.usd?.unit_amount === want.usd &&
		price.currency_options?.eur?.unit_amount === want.eur
	);
}

async function ensureProduct(stripe: Stripe, product: StripeProduct): Promise<string> {
	const id = `hexascore_${product}`;
	try {
		const existing = await stripe.products.retrieve(id);
		if (!existing.active || existing.name !== NAMES[product]) {
			await stripe.products.update(id, { active: true, name: NAMES[product] });
		}
		return id;
	} catch (error) {
		if ((error as { statusCode?: number }).statusCode !== 404) throw error;
		await stripe.products.create({ id, name: NAMES[product], metadata: { app: "hexascore" } });
		return id;
	}
}

async function ensurePrice(stripe: Stripe, product: StripeProduct): Promise<Stripe.Price> {
	const lookupKey = STRIPE_LOOKUP_KEYS[product];
	const productId = await ensureProduct(stripe, product);
	const [current] = (
		await stripe.prices.list({ lookup_keys: [lookupKey], expand: ["data.currency_options"], limit: 1 })
	).data;
	if (current && matches(current, product)) return current;

	const want = amounts(product);
	const created = await stripe.prices.create({
		product: productId,
		currency: "brl",
		unit_amount: want.brl,
		currency_options: { usd: { unit_amount: want.usd }, eur: { unit_amount: want.eur } },
		...(product === "credits" ? {} : { recurring: { interval: "month" as const } }),
		lookup_key: lookupKey,
		transfer_lookup_key: true,
		metadata: { app: "hexascore" },
	});
	if (current) await stripe.prices.update(current.id, { active: false });
	return created;
}

async function ensurePortal(stripe: Stripe, prices: Record<"pro" | "business", Stripe.Price>): Promise<string> {
	const products = (["pro", "business"] as const).map((p) => ({
		product: typeof prices[p].product === "string" ? prices[p].product : prices[p].product.id,
		prices: [prices[p].id],
	}));
	const features: Stripe.BillingPortal.ConfigurationCreateParams.Features = {
		customer_update: { enabled: true, allowed_updates: ["email", "address", "tax_id"] },
		invoice_history: { enabled: true },
		payment_method_update: { enabled: true },
		subscription_cancel: { enabled: true, mode: "at_period_end" },
		subscription_update: {
			enabled: true,
			default_allowed_updates: ["price"],
			products,
			// Upgrades charge the prorated difference now; downgrades wait for the period end.
			proration_behavior: "always_invoice",
			schedule_at_period_end: { conditions: [{ type: "decreasing_item_amount" }] },
		},
	};
	const existing = (await stripe.billingPortal.configurations.list({ limit: 100 })).data.find(
		(c) => c.metadata?.app === PORTAL_CONFIG_METADATA.app,
	);
	if (existing) {
		await stripe.billingPortal.configurations.update(existing.id, { active: true, features });
		return existing.id;
	}
	const created = await stripe.billingPortal.configurations.create({
		features,
		metadata: { ...PORTAL_CONFIG_METADATA },
	});
	return created.id;
}

if (import.meta.main) {
	const stripe = getStripe();
	if (!stripe) {
		console.error("[stripe-setup] STRIPE_SECRET_KEY is not set; nothing to do.");
		process.exit(1);
	}
	try {
		const pro = await ensurePrice(stripe, "pro");
		const business = await ensurePrice(stripe, "business");
		const credits = await ensurePrice(stripe, "credits");
		const portal = await ensurePortal(stripe, { pro, business });
		for (const p of [pro, business, credits]) console.info(`[stripe-setup] ${p.lookup_key} → ${p.id}`);
		console.info(`[stripe-setup] portal configuration → ${portal}`);
		process.exit(0);
	} catch (error) {
		console.error("[stripe-setup] failed:", error);
		process.exit(1);
	}
}
