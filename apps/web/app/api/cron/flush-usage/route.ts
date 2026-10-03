import { sql } from "drizzle-orm";
import { creditsUsedKey, dayId, getStore, usageKey } from "@/lib/api/quota";
import { isAuthorizedCron } from "@/lib/cron";
import { db } from "@/lib/db/client";
import { apiUsageDaily, creditLedger } from "@/lib/db/schema";

// Idempotent: counts are written as absolute values and credit debits use a unique ref per user/day.
export async function GET(request: Request): Promise<Response> {
	if (!isAuthorizedCron(request)) return new Response("Unauthorized", { status: 401 });
	const store = getStore();
	if (!store) return Response.json({ error: "rate-limit store not configured" }, { status: 503 });

	const days = [dayId(new Date(Date.now() - 86_400_000)), dayId()];
	const summary: Record<string, { users: number; creditDebits: number }> = {};
	try {
		for (const day of days) {
			const usage = await store.hgetall(usageKey(day));
			const credits = await store.hgetall(creditsUsedKey(day));
			const usageRows = Object.entries(usage)
				.filter(([, n]) => n > 0)
				.map(([userId, count]) => ({ userId, day, count }));
			if (usageRows.length) {
				await db
					.insert(apiUsageDaily)
					.values(usageRows)
					.onConflictDoUpdate({
						target: [apiUsageDaily.userId, apiUsageDaily.day],
						set: { count: sql`excluded.count` },
					});
			}
			// Only closed days are debited, so a credit spent today is never written twice with different totals.
			const creditRows =
				day === dayId()
					? []
					: Object.entries(credits)
							.filter(([, n]) => n > 0)
							.map(([userId, n]) => ({
								userId,
								delta: -n,
								reason: "api_usage",
								ref: `usage:${day}:${userId}`,
							}));
			if (creditRows.length) await db.insert(creditLedger).values(creditRows).onConflictDoNothing();
			summary[day] = { users: usageRows.length, creditDebits: creditRows.length };
		}
		console.info("[cron] flush-usage", JSON.stringify(summary));
		return Response.json({ ok: true, summary });
	} catch (error) {
		console.error("[cron] flush-usage failed", error);
		return Response.json({ ok: false }, { status: 500 });
	}
}
