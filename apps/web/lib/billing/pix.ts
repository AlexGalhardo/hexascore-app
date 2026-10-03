import { and, eq, ne } from "drizzle-orm";
import type { CounterStore } from "../api/quota";
import type { Database } from "../db/client";
import { pixCharges, processedEvents, subscriptions } from "../db/schema";
import {
	type AbacateWebhook,
	abacateWebhookSchema,
	chargeRefFromWebhook,
	safeEqual,
	verifyAbacateSignature,
} from "./abacatepay";
import { extendPeriod, type PixPurpose, pixOffer, pixPurposeSchema } from "./catalog";
import {
	type Executor,
	getSubscription,
	insertCreditPurchase,
	isActive,
	mirrorCreditsToCache,
	type SubscriptionRow,
} from "./subscriptions";

export type PixChargeRow = typeof pixCharges.$inferSelect;

/** Why a PIX purchase is refused for this user right now, or null when allowed. */
export function pixRefusal(
	purpose: PixPurpose,
	sub: Pick<SubscriptionRow, "status" | "currentPeriodEnd" | "provider" | "plan"> | null,
	now = new Date(),
): "stripe_active" | "plan_conflict" | "plan_required" | null {
	const offer = pixOffer(purpose);
	const active = isActive(sub, now);
	if (offer.kind === "credits") return active && sub?.plan !== "free" ? null : "plan_required";
	if (!active || !sub) return null;
	if (sub.provider === "stripe") return "stripe_active";
	return sub.plan === offer.plan ? null : "plan_conflict";
}

/**
 * Applies a paid PIX charge exactly once (the pending → paid transition is the lock).
 * Plans stack on the current PIX period; credits go to the ledger with ref `pix:<chargeId>`.
 */
export async function applyPixPayment(
	tx: Executor,
	chargeId: string,
	now = new Date(),
): Promise<{ applied: boolean; creditsFor?: { userId: string; requests: number } }> {
	const [charge] = await tx
		.update(pixCharges)
		.set({ status: "paid", paidAt: now })
		.where(and(eq(pixCharges.id, chargeId), ne(pixCharges.status, "paid")))
		.returning();
	if (!charge) return { applied: false };

	const purpose = pixPurposeSchema.parse(charge.purpose);
	const offer = pixOffer(purpose);
	if (offer.kind === "credits") {
		const inserted = await insertCreditPurchase(tx, charge.userId, offer.requests, `pix:${charge.id}`);
		return inserted
			? { applied: true, creditsFor: { userId: charge.userId, requests: offer.requests } }
			: { applied: true };
	}

	const existing = await getSubscription(tx, charge.userId);
	if (existing?.provider === "stripe" && isActive(existing, now)) {
		// Creation refuses this; reaching here means a race. Never clobber a card subscription: flag for refund.
		console.error(
			`[pix] charge ${charge.id} paid while ${charge.userId} has an active Stripe subscription: refund manually`,
		);
		return { applied: true };
	}
	const stacks = existing?.provider === "abacatepay" && existing.plan === offer.plan;
	const currentPeriodEnd = extendPeriod(stacks ? existing.currentPeriodEnd : null, now, offer.months);
	const values = {
		plan: offer.plan,
		provider: "abacatepay" as const,
		status: "active" as const,
		currentPeriodEnd,
		cancelAtPeriodEnd: false,
		providerCustomerId: null,
		providerRef: charge.providerRef,
	};
	await tx
		.insert(subscriptions)
		.values({ userId: charge.userId, ...values })
		.onConflictDoUpdate({ target: subscriptions.userId, set: values });
	return { applied: true };
}

export async function settlePixCharge(
	db: Database,
	store: CounterStore | null,
	chargeId: string,
	now = new Date(),
): Promise<boolean> {
	const result = await db.transaction((tx) => applyPixPayment(tx, chargeId, now));
	if (result.creditsFor) await mirrorCreditsToCache(store, result.creditsFor.userId, result.creditsFor.requests);
	return result.applied;
}

export type AbacateWebhookDeps = { db: Database; store: CounterStore | null; secret: string; now?: () => Date };

/** 401 = not from AbacatePay, 200 = applied/duplicate/ignored, 500 = transient (AbacatePay retries). */
export async function handleAbacateWebhook(
	rawBody: string,
	query: { webhookSecret: string | null; signature: string | null },
	deps: AbacateWebhookDeps,
): Promise<{ status: number; body: Record<string, unknown> }> {
	if (!query.webhookSecret || !safeEqual(query.webhookSecret, deps.secret)) {
		console.warn("[pix] webhook rejected: bad webhookSecret");
		return { status: 401, body: { error: "unauthorized" } };
	}
	if (!verifyAbacateSignature(rawBody, query.signature)) {
		console.warn("[pix] webhook rejected: bad X-Webhook-Signature");
		return { status: 401, body: { error: "invalid signature" } };
	}
	let payload: AbacateWebhook;
	try {
		payload = abacateWebhookSchema.parse(JSON.parse(rawBody));
	} catch (error) {
		console.error("[pix] webhook payload rejected", error);
		return { status: 400, body: { error: "invalid payload" } };
	}
	if (payload.event !== "transparent.completed")
		return { status: 200, body: { received: true, ignored: payload.event } };

	const ref = chargeRefFromWebhook(payload);
	if (!ref) return { status: 400, body: { error: "missing charge id" } };
	const now = deps.now?.() ?? new Date();
	try {
		const outcome = await deps.db.transaction(async (tx) => {
			const claimed = await tx
				.insert(processedEvents)
				.values({ id: payload.id, provider: "abacatepay" })
				.onConflictDoNothing()
				.returning({ id: processedEvents.id });
			if (!claimed.length) return { note: "duplicate" } as const;
			const [charge] = await tx.select().from(pixCharges).where(eq(pixCharges.providerRef, ref.id));
			if (!charge) return { note: `unknown charge ${ref.id}` } as const;
			if (ref.amount !== null && ref.amount !== charge.amountCents) {
				console.error(
					`[pix] amount mismatch on ${charge.id}: paid ${ref.amount}, expected ${charge.amountCents}`,
				);
				return { note: "amount mismatch" } as const;
			}
			const result = await applyPixPayment(tx, charge.id, now);
			return { note: result.applied ? "applied" : "already applied", creditsFor: result.creditsFor } as const;
		});
		if ("creditsFor" in outcome && outcome.creditsFor) {
			await mirrorCreditsToCache(deps.store, outcome.creditsFor.userId, outcome.creditsFor.requests);
		}
		console.info(`[pix] webhook ${payload.id} (${ref.id}): ${outcome.note}`);
		return { status: 200, body: { received: true } };
	} catch (error) {
		console.error(`[pix] failed to process webhook ${payload.id}`, error);
		return { status: 500, body: { error: "processing failed" } };
	}
}
