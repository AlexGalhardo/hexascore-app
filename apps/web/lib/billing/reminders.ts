import { and, eq, gt, inArray, lte } from "drizzle-orm";
import type { Database } from "../db/client";
import { processedEvents, subscriptions, user } from "../db/schema";

export const REMINDER_DAYS = [7, 1] as const;
const DAY_MS = 86_400_000;

export type Reminder = { userId: string; email: string; plan: string; currentPeriodEnd: Date; days: number };

/** Whole days left, rounded up: a period ending in 6d 3h is "7 days". */
export function daysLeft(end: Date, now: Date): number {
	return Math.ceil((end.getTime() - now.getTime()) / DAY_MS);
}

/** PIX subscriptions whose period ends in exactly 7 or 1 day(s), each claimed once per period end. */
export async function dueReminders(db: Database, now: Date): Promise<Reminder[]> {
	const rows = await db
		.select({
			userId: subscriptions.userId,
			email: user.email,
			plan: subscriptions.plan,
			currentPeriodEnd: subscriptions.currentPeriodEnd,
		})
		.from(subscriptions)
		.innerJoin(user, eq(user.id, subscriptions.userId))
		.where(
			and(
				eq(subscriptions.provider, "abacatepay"),
				eq(subscriptions.status, "active"),
				gt(subscriptions.currentPeriodEnd, now),
				lte(subscriptions.currentPeriodEnd, new Date(now.getTime() + Math.max(...REMINDER_DAYS) * DAY_MS)),
			),
		);
	return rows
		.map((r) => ({ ...r, days: daysLeft(r.currentPeriodEnd, now) }))
		.filter((r) => (REMINDER_DAYS as readonly number[]).includes(r.days));
}

/** processed_events doubles as the "already emailed" log, so a cron re-run never sends twice. */
export async function claimReminder(db: Database, r: Reminder): Promise<boolean> {
	const rows = await db
		.insert(processedEvents)
		.values({ id: `reminder:${r.userId}:${r.currentPeriodEnd.toISOString()}:${r.days}`, provider: "abacatepay" })
		.onConflictDoNothing()
		.returning({ id: processedEvents.id });
	return rows.length > 0;
}

export async function releaseReminder(db: Database, r: Reminder): Promise<void> {
	await db
		.delete(processedEvents)
		.where(eq(processedEvents.id, `reminder:${r.userId}:${r.currentPeriodEnd.toISOString()}:${r.days}`));
}

/** PIX periods are not renewed by a provider, so the cron is what ends them. */
export async function expireSubscriptions(db: Database, now: Date): Promise<number> {
	const rows = await db
		.update(subscriptions)
		.set({ status: "expired" })
		.where(
			and(
				eq(subscriptions.provider, "abacatepay"),
				inArray(subscriptions.status, ["active", "past_due"]),
				lte(subscriptions.currentPeriodEnd, now),
			),
		)
		.returning({ userId: subscriptions.userId });
	return rows.length;
}
