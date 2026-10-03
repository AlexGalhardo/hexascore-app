import { and, eq, inArray, isNotNull, isNull, lt, sql } from "drizzle-orm";
import { purgeCutoff } from "./account-deletion";
import type { Database } from "./db/client";
import { account, apiKeys, passkey, session, subscriptions, twoFactor, user, verification } from "./db/schema";

export type CancelStripeSubscription = (subscriptionId: string) => Promise<void>;

// Anonymizes instead of deleting the row: ratings keep their user_id so the metascore is not distorted,
// but nothing left on the row identifies the person and no credential (session, password, API key) works again.
// A live Stripe subscription is canceled first; if that fails (or Stripe is not configured) the account stays
// scheduled and the next daily run retries, so a deleted account is never billed again.
export async function purgeScheduledAccounts(
	db: Database,
	now = new Date(),
	cancelStripe?: CancelStripeSubscription,
): Promise<string[]> {
	const due = await db
		.select({ id: user.id, email: user.email })
		.from(user)
		.where(and(isNotNull(user.deletionScheduledAt), lt(user.deletionScheduledAt, purgeCutoff(now))));
	if (due.length === 0) return [];

	const liveStripe = await db
		.select({ userId: subscriptions.userId, ref: subscriptions.providerRef })
		.from(subscriptions)
		.where(
			and(
				inArray(
					subscriptions.userId,
					due.map((u) => u.id),
				),
				eq(subscriptions.provider, "stripe"),
				inArray(subscriptions.status, ["active", "past_due"]),
				isNotNull(subscriptions.providerRef),
			),
		);
	const blocked = new Set<string>();
	for (const sub of liveStripe) {
		try {
			if (!cancelStripe || !sub.ref) throw new Error("Stripe is not configured");
			await cancelStripe(sub.ref);
		} catch (error) {
			blocked.add(sub.userId);
			console.error(
				`[purge-accounts] could not cancel Stripe subscription of ${sub.userId}, retrying tomorrow:`,
				error,
			);
		}
	}

	const ready = due.filter((u) => !blocked.has(u.id));
	if (ready.length === 0) return [];
	const ids = ready.map((u) => u.id);

	await db.transaction(async (tx) => {
		await tx.delete(session).where(inArray(session.userId, ids));
		await tx.delete(account).where(inArray(account.userId, ids));
		await tx.delete(passkey).where(inArray(passkey.userId, ids));
		await tx.delete(twoFactor).where(inArray(twoFactor.userId, ids));
		await tx.delete(verification).where(
			inArray(
				verification.identifier,
				ready.map((u) => u.email),
			),
		);
		await tx
			.update(apiKeys)
			.set({ revokedAt: now })
			.where(and(inArray(apiKeys.userId, ids), isNull(apiKeys.revokedAt)));
		await tx
			.update(subscriptions)
			.set({ status: "canceled", cancelAtPeriodEnd: false })
			.where(and(inArray(subscriptions.userId, ids), inArray(subscriptions.status, ["active", "past_due"])));
		await tx
			.update(user)
			.set({
				email: sql`'deleted-' || ${user.id} || '@invalid'`,
				name: "Deleted user",
				username: null,
				displayUsername: null,
				image: null,
				emailVerified: false,
				twoFactorEnabled: false,
				profilePrivate: true,
				deletionScheduledAt: null,
			})
			.where(inArray(user.id, ids));
	});
	return ids;
}
