import { eq, sql } from "drizzle-orm";
import { addCachedCredits, type CounterStore, creditsCacheKey } from "../api/quota";
import type { Database } from "../db/client";
import { creditLedger, stripeCustomers, subscriptions } from "../db/schema";
import type { PaidPlan } from "./catalog";
import type { NormalizedSubscription } from "./stripe";

export type Tx = Parameters<Parameters<Database["transaction"]>[0]>[0];
export type Executor = Database | Tx;
export type SubscriptionRow = typeof subscriptions.$inferSelect;

export function isActive(row: Pick<SubscriptionRow, "status" | "currentPeriodEnd"> | null, now = new Date()): boolean {
	return Boolean(row && row.status === "active" && row.currentPeriodEnd.getTime() > now.getTime());
}

export async function getSubscription(ex: Executor, userId: string): Promise<SubscriptionRow | null> {
	const [row] = await ex.select().from(subscriptions).where(eq(subscriptions.userId, userId));
	return row ?? null;
}

export async function creditBalance(ex: Executor, userId: string): Promise<number> {
	const [row] = await ex
		.select({ balance: sql<string>`coalesce(sum(${creditLedger.delta}), 0)` })
		.from(creditLedger)
		.where(eq(creditLedger.userId, userId));
	return Number(row?.balance ?? 0);
}

/** Inserts a purchase into the ledger once per `ref`; returns false when it was already recorded. */
export async function insertCreditPurchase(
	ex: Executor,
	userId: string,
	requests: number,
	ref: string,
): Promise<boolean> {
	const rows = await ex
		.insert(creditLedger)
		.values({ userId, delta: requests, reason: "purchase", ref })
		.onConflictDoNothing({ target: creditLedger.ref })
		.returning({ id: creditLedger.id });
	return rows.length > 0;
}

/** After the ledger commit: bump the Redis balance if it is cached. Failure only delays credits, so it is logged. */
export async function mirrorCreditsToCache(
	store: CounterStore | null,
	userId: string,
	requests: number,
): Promise<void> {
	if (!store) return;
	try {
		await addCachedCredits(store, userId, requests);
	} catch (error) {
		console.error(`[billing] could not add ${requests} cached credits for ${userId}; dropping the key`, error);
		await store.del(creditsCacheKey(userId)).catch(() => undefined);
	}
}

export async function userIdForCustomer(ex: Executor, customerId: string): Promise<string | null> {
	const [row] = await ex
		.select({ userId: stripeCustomers.userId })
		.from(stripeCustomers)
		.where(eq(stripeCustomers.customerId, customerId));
	return row?.userId ?? null;
}

export async function linkCustomer(ex: Executor, userId: string, customerId: string): Promise<void> {
	await ex.insert(stripeCustomers).values({ userId, customerId }).onConflictDoNothing();
}

export type ApplyResult = "applied" | "skipped";

/**
 * Writes Stripe's view of a subscription. An active PIX period or a newer Stripe subscription is never
 * overwritten by a non-active state of some other subscription (e.g. an old one finally canceled).
 */
export async function applyStripeSubscription(
	ex: Executor,
	userId: string,
	sub: NormalizedSubscription & { plan: PaidPlan },
	now = new Date(),
): Promise<ApplyResult> {
	const existing = await getSubscription(ex, userId);
	const sameSubscription = existing?.provider === "stripe" && existing.providerRef === sub.subscriptionId;
	if (existing && !sameSubscription && isActive(existing, now) && sub.status !== "active") return "skipped";
	if (existing?.provider === "abacatepay" && isActive(existing, now)) {
		console.warn(`[billing] stripe subscription ${sub.subscriptionId} replaces an active PIX period of ${userId}`);
	}
	const values = {
		plan: sub.plan,
		provider: "stripe" as const,
		status: sub.status,
		currentPeriodEnd: sub.currentPeriodEnd,
		cancelAtPeriodEnd: sub.cancelAtPeriodEnd,
		providerCustomerId: sub.customerId,
		providerRef: sub.subscriptionId,
	};
	await ex
		.insert(subscriptions)
		.values({ userId, ...values })
		.onConflictDoUpdate({ target: subscriptions.userId, set: values });
	return "applied";
}
