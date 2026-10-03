"use server";

import { and, desc, eq, gt } from "drizzle-orm";
import { getLocale } from "next-intl/server";
import { z } from "zod";
import { getPathname } from "@/i18n/navigation";
import { routing } from "@/i18n/routing";
import { createPixQrCode, simulatePixPayment } from "@/lib/billing/abacatepay";
import { currencySchema, pixOffer, pixPurposeSchema, STRIPE_LOOKUP_KEYS } from "@/lib/billing/catalog";
import { pixRefusal, simulateSandboxPayment } from "@/lib/billing/pix";
import { getStripe, portalConfigurationId, priceIdFor } from "@/lib/billing/stripe";
import { getSubscription, isActive, linkCustomer } from "@/lib/billing/subscriptions";
import { APP_NAME, APP_URL } from "@/lib/config";
import { db } from "@/lib/db/client";
import { pixCharges, stripeCustomers } from "@/lib/db/schema";
import { features, pixSandbox } from "@/lib/env";
import { pixCheckoutFlag } from "@/lib/flags";
import { getSessionUser, type SessionUser } from "@/lib/session";

export type BillingError =
	| "unauthenticated"
	| "unverified"
	| "not_configured"
	| "invalid"
	| "already_subscribed"
	| "pix_active"
	| "stripe_active"
	| "plan_conflict"
	| "plan_required"
	| "no_customer"
	| "not_found"
	| "not_pending"
	| "server";

export type RedirectResult = { ok: true; url: string } | { ok: false; error: BillingError };

const PIX_EXPIRES_SECONDS = 60 * 60;

async function payer(): Promise<SessionUser | BillingError> {
	const user = await getSessionUser();
	if (!user) return "unauthenticated";
	if (!user.emailVerified) return "unverified";
	return user;
}

async function checkoutUrl(query: string): Promise<string> {
	const requested = await getLocale();
	const locale = routing.locales.find((l) => l === requested) ?? routing.defaultLocale;
	return `${APP_URL}${getPathname({ href: "/profile/api/checkout", locale })}${query}`;
}

const checkoutInput = z.object({
	product: z.enum(["pro", "business", "credits"]),
	currency: currencySchema,
});

/** Creates a Stripe Checkout Session; the price comes from Stripe by lookup_key, never from the client. */
export async function startStripeCheckout(input: z.input<typeof checkoutInput>): Promise<RedirectResult> {
	const user = await payer();
	if (typeof user === "string") return { ok: false, error: user };
	const stripe = getStripe();
	if (!stripe) return { ok: false, error: "not_configured" };
	const parsed = checkoutInput.safeParse(input);
	if (!parsed.success) return { ok: false, error: "invalid" };
	const { product, currency } = parsed.data;

	try {
		const sub = await getSubscription(db, user.id);
		const active = isActive(sub);
		if (product === "credits") {
			if (!active || sub?.plan === "free") return { ok: false, error: "plan_required" };
		} else if (active && sub?.provider === "stripe") {
			return { ok: false, error: "already_subscribed" };
		} else if (active && sub?.provider === "abacatepay") {
			return { ok: false, error: "pix_active" };
		}

		const price = await priceIdFor(stripe, product);
		if (!price) {
			console.error(`[billing] Stripe price ${STRIPE_LOOKUP_KEYS[product]} missing: run scripts/stripe-setup.ts`);
			return { ok: false, error: "not_configured" };
		}

		let [customer] = await db.select().from(stripeCustomers).where(eq(stripeCustomers.userId, user.id));
		if (!customer) {
			const created = await stripe.customers.create(
				{ email: user.email, name: user.name, metadata: { user_id: user.id } },
				{ idempotencyKey: `customer-${user.id}` },
			);
			await linkCustomer(db, user.id, created.id);
			[customer] = await db.select().from(stripeCustomers).where(eq(stripeCustomers.userId, user.id));
		}
		if (!customer) return { ok: false, error: "server" };

		const kind = product === "credits" ? "credits" : "subscription";
		const metadata = { user_id: user.id, kind };
		const locale = await getLocale();
		const session = await stripe.checkout.sessions.create({
			mode: kind === "credits" ? "payment" : "subscription",
			line_items: [{ price, quantity: 1 }],
			currency,
			customer: customer.customerId,
			client_reference_id: user.id,
			metadata,
			...(kind === "credits"
				? { payment_intent_data: { metadata } }
				: { subscription_data: { metadata: { user_id: user.id } } }),
			locale: locale === "pt-BR" ? "pt-BR" : locale === "es" ? "es" : "en",
			success_url: await checkoutUrl("?status=success"),
			cancel_url: await checkoutUrl("?status=canceled"),
		});
		if (!session.url) return { ok: false, error: "server" };
		return { ok: true, url: session.url };
	} catch (error) {
		console.error("[billing] stripe checkout failed", error);
		return { ok: false, error: "server" };
	}
}

export async function openBillingPortal(): Promise<RedirectResult> {
	const user = await getSessionUser();
	if (!user) return { ok: false, error: "unauthenticated" };
	const stripe = getStripe();
	if (!stripe) return { ok: false, error: "not_configured" };
	try {
		const [customer] = await db.select().from(stripeCustomers).where(eq(stripeCustomers.userId, user.id));
		if (!customer) return { ok: false, error: "no_customer" };
		const session = await stripe.billingPortal.sessions.create({
			customer: customer.customerId,
			configuration: await portalConfigurationId(stripe),
			return_url: await checkoutUrl(""),
		});
		return { ok: true, url: session.url };
	} catch (error) {
		console.error("[billing] portal session failed", error);
		return { ok: false, error: "server" };
	}
}

export type PixChargeView = {
	id: string;
	brCode: string;
	qrImage: string;
	amountCents: number;
	expiresAt: string;
};

export type PixResult = { ok: true; charge: PixChargeView } | { ok: false; error: BillingError };

function toView(row: typeof pixCharges.$inferSelect): PixChargeView | null {
	if (!row.brCode || !row.brCodeBase64 || !row.expiresAt) return null;
	const qrImage = row.brCodeBase64.startsWith("data:")
		? row.brCodeBase64
		: `data:image/png;base64,${row.brCodeBase64}`;
	return {
		id: row.id,
		brCode: row.brCode,
		qrImage,
		amountCents: row.amountCents,
		expiresAt: row.expiresAt.toISOString(),
	};
}

/** PIX (BRL only): the amount comes from the catalog for the chosen purpose; a pending code is reused. */
export async function createPixCharge(input: { purpose: string }): Promise<PixResult> {
	const user = await payer();
	if (typeof user === "string") return { ok: false, error: user };
	if (!features.pix || !(await pixCheckoutFlag())) return { ok: false, error: "not_configured" };
	const parsed = pixPurposeSchema.safeParse(input?.purpose);
	if (!parsed.success) return { ok: false, error: "invalid" };
	const purpose = parsed.data;

	try {
		const refusal = pixRefusal(purpose, await getSubscription(db, user.id));
		if (refusal) return { ok: false, error: refusal };

		const now = new Date();
		const [pending] = await db
			.select()
			.from(pixCharges)
			.where(
				and(
					eq(pixCharges.userId, user.id),
					eq(pixCharges.purpose, purpose),
					eq(pixCharges.status, "pending"),
					gt(pixCharges.expiresAt, new Date(now.getTime() + 5 * 60_000)),
				),
			)
			.orderBy(desc(pixCharges.createdAt))
			.limit(1);
		const reusable = pending && toView(pending);
		if (reusable) return { ok: true, charge: reusable };

		const offer = pixOffer(purpose);
		const qr = await createPixQrCode({
			amountCents: offer.amountCents,
			description: `${APP_NAME} API — ${purpose}`,
			expiresInSeconds: PIX_EXPIRES_SECONDS,
			metadata: { userId: user.id, purpose },
		});
		if (qr.amount !== offer.amountCents) {
			console.error(
				`[billing] AbacatePay returned amount ${qr.amount} for ${purpose} (expected ${offer.amountCents})`,
			);
			return { ok: false, error: "server" };
		}
		const [row] = await db
			.insert(pixCharges)
			.values({
				userId: user.id,
				providerRef: qr.id,
				purpose,
				amountCents: offer.amountCents,
				brCode: qr.brCode,
				brCodeBase64: qr.brCodeBase64,
				expiresAt: qr.expiresAt,
			})
			.returning();
		const view = row && toView(row);
		return view ? { ok: true, charge: view } : { ok: false, error: "server" };
	} catch (error) {
		console.error("[billing] pix charge failed", error);
		return { ok: false, error: "server" };
	}
}

export type SimulateResult = { ok: true } | { ok: false; error: BillingError };

/** Sandbox testing only (ABACATEPAY_SANDBOX=true): asks AbacatePay to pay the user's own pending PIX charge. */
export async function simulatePixPaymentAction(chargeId: string): Promise<SimulateResult> {
	const user = await payer();
	if (typeof user === "string") return { ok: false, error: user };
	if (!pixSandbox) return { ok: false, error: "not_configured" };
	const parsed = z.uuid().safeParse(chargeId);
	if (!parsed.success) return { ok: false, error: "invalid" };
	try {
		const outcome = await simulateSandboxPayment(user.id, parsed.data, {
			db,
			enabled: pixSandbox,
			simulate: simulatePixPayment,
		});
		return outcome === "ok" ? { ok: true } : { ok: false, error: outcome };
	} catch (error) {
		console.error("[billing] sandbox pix simulation failed", error);
		return { ok: false, error: "server" };
	}
}
