import { eq } from "drizzle-orm";
import { db } from "../db/client";
import { stripeCustomers } from "../db/schema";
import { creditBalance, getSubscription, isActive } from "./subscriptions";

/** Browser-safe DTO for the billing page: no provider ids leave the server. */
export type BillingSummary = {
	plan: "free" | "pro" | "business";
	status: "active" | "past_due" | "canceled" | "expired" | null;
	provider: "stripe" | "abacatepay" | null;
	currentPeriodEnd: string | null;
	cancelAtPeriodEnd: boolean;
	credits: number;
	hasCardBilling: boolean;
};

export async function getBillingSummary(userId: string): Promise<BillingSummary> {
	const [sub, credits, [customer]] = await Promise.all([
		getSubscription(db, userId),
		creditBalance(db, userId),
		db.select({ id: stripeCustomers.customerId }).from(stripeCustomers).where(eq(stripeCustomers.userId, userId)),
	]);
	return {
		plan: isActive(sub) && sub ? sub.plan : "free",
		status: sub?.status ?? null,
		provider: sub?.provider ?? null,
		currentPeriodEnd: sub?.currentPeriodEnd.toISOString() ?? null,
		cancelAtPeriodEnd: sub?.cancelAtPeriodEnd ?? false,
		credits,
		hasCardBilling: Boolean(customer),
	};
}
