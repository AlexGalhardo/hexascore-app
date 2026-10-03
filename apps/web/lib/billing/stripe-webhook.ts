import { eq } from "drizzle-orm";
import type Stripe from "stripe";
import { z } from "zod";
import { CREDIT_PACK_REQUESTS } from "../api/plans";
import type { CounterStore } from "../api/quota";
import type { Database } from "../db/client";
import { processedEvents, user } from "../db/schema";
import { normalizeSubscription } from "./stripe";
import {
	applyStripeSubscription,
	type Executor,
	insertCreditPurchase,
	linkCustomer,
	mirrorCreditsToCache,
	userIdForCustomer,
} from "./subscriptions";

export type StripeWebhookDeps = {
	stripe: Stripe;
	secret: string;
	db: Database;
	store: CounterStore | null;
	/** Fetches the current subscription so out-of-order events never write stale state. */
	retrieveSubscription: (id: string) => Promise<Stripe.Subscription>;
	now?: () => Date;
};

export type WebhookResult = { status: number; body: Record<string, unknown> };

const checkoutMetadata = z.object({
	user_id: z.string().min(1),
	kind: z.enum(["subscription", "credits"]),
});

const HANDLED = new Set<string>([
	"checkout.session.completed",
	"checkout.session.async_payment_succeeded",
	"customer.subscription.created",
	"customer.subscription.updated",
	"customer.subscription.deleted",
	"invoice.paid",
	"invoice.payment_failed",
]);

type Effect = (tx: Executor) => Promise<{ creditsFor?: { userId: string; requests: number }; note: string }>;

async function userExists(ex: Executor, id: string): Promise<boolean> {
	const [row] = await ex.select({ id: user.id }).from(user).where(eq(user.id, id));
	return Boolean(row);
}

function idOf(ref: string | { id: string } | null | undefined): string | null {
	if (!ref) return null;
	return typeof ref === "string" ? ref : ref.id;
}

async function subscriptionEffect(deps: StripeWebhookDeps, subscriptionId: string): Promise<Effect> {
	const sub = normalizeSubscription(await deps.retrieveSubscription(subscriptionId));
	return async (tx) => {
		const userId = sub.metadataUserId ?? (await userIdForCustomer(tx, sub.customerId));
		if (!userId || !(await userExists(tx, userId))) return { note: `no local user for ${sub.subscriptionId}` };
		if (!sub.plan) {
			console.error(`[stripe] subscription ${sub.subscriptionId} has no hexascore price lookup_key`);
			return { note: "unknown price" };
		}
		await linkCustomer(tx, userId, sub.customerId);
		const result = await applyStripeSubscription(tx, userId, { ...sub, plan: sub.plan }, deps.now?.());
		return { note: `subscription ${result}` };
	};
}

async function effectFor(deps: StripeWebhookDeps, event: Stripe.Event): Promise<Effect | null> {
	switch (event.type) {
		case "checkout.session.completed":
		case "checkout.session.async_payment_succeeded": {
			const session = event.data.object;
			const meta = checkoutMetadata.safeParse(session.metadata ?? {});
			if (!meta.success) return null;
			if (meta.data.kind === "subscription") {
				const subscriptionId = idOf(session.subscription);
				return subscriptionId ? subscriptionEffect(deps, subscriptionId) : null;
			}
			if (session.payment_status !== "paid") return async () => ({ note: "credits awaiting payment" });
			const userId = meta.data.user_id;
			return async (tx) => {
				if (!(await userExists(tx, userId))) return { note: "user gone" };
				const customerId = idOf(session.customer);
				if (customerId) await linkCustomer(tx, userId, customerId);
				const inserted = await insertCreditPurchase(tx, userId, CREDIT_PACK_REQUESTS, session.id);
				return inserted
					? { creditsFor: { userId, requests: CREDIT_PACK_REQUESTS }, note: "credits granted" }
					: { note: "credits already granted" };
			};
		}
		case "customer.subscription.created":
		case "customer.subscription.updated":
		case "customer.subscription.deleted":
			return subscriptionEffect(deps, event.data.object.id);
		case "invoice.paid":
		case "invoice.payment_failed": {
			const invoice = event.data.object;
			const subscriptionId = idOf(invoice.parent?.subscription_details?.subscription);
			if (event.type === "invoice.payment_failed") {
				console.warn(
					`[stripe] invoice ${invoice.id} payment failed (subscription ${subscriptionId ?? "none"})`,
				);
			}
			return subscriptionId ? subscriptionEffect(deps, subscriptionId) : null;
		}
		default:
			return null;
	}
}

/**
 * Verifies, deduplicates and applies one Stripe webhook delivery.
 * 400 = bad signature (Stripe stops), 200 = applied/duplicate/ignored, 500 = transient (Stripe retries).
 */
export async function handleStripeWebhook(
	rawBody: string,
	signature: string | null,
	deps: StripeWebhookDeps,
): Promise<WebhookResult> {
	if (!signature) return { status: 400, body: { error: "missing stripe-signature header" } };
	let event: Stripe.Event;
	try {
		event = await deps.stripe.webhooks.constructEventAsync(rawBody, signature, deps.secret);
	} catch (error) {
		console.warn("[stripe] webhook signature verification failed", error instanceof Error ? error.message : error);
		return { status: 400, body: { error: "invalid signature" } };
	}
	if (!HANDLED.has(event.type)) return { status: 200, body: { received: true, ignored: event.type } };

	try {
		const [seen] = await deps.db
			.select({ id: processedEvents.id })
			.from(processedEvents)
			.where(eq(processedEvents.id, event.id));
		if (seen) return { status: 200, body: { received: true, duplicate: true } };

		const effect = await effectFor(deps, event);
		const outcome = await deps.db.transaction(async (tx) => {
			const claimed = await tx
				.insert(processedEvents)
				.values({ id: event.id, provider: "stripe" })
				.onConflictDoNothing()
				.returning({ id: processedEvents.id });
			if (!claimed.length) return { duplicate: true as const };
			return { duplicate: false as const, ...(effect ? await effect(tx) : { note: "nothing to do" }) };
		});
		if (outcome.duplicate) return { status: 200, body: { received: true, duplicate: true } };
		if (outcome.creditsFor) {
			await mirrorCreditsToCache(deps.store, outcome.creditsFor.userId, outcome.creditsFor.requests);
		}
		console.info(`[stripe] ${event.type} ${event.id}: ${outcome.note}`);
		return { status: 200, body: { received: true } };
	} catch (error) {
		console.error(`[stripe] failed to process ${event.type} ${event.id}`, error);
		return { status: 500, body: { error: "processing failed" } };
	}
}
