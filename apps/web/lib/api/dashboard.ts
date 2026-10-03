import { and, desc, eq, gte } from "drizzle-orm";
import type { Database } from "../db/client";
import { apiKeys, apiUsageDaily } from "../db/schema";
import { PLAN_LIMITS, type PlanId } from "./plans";
import { type CounterStore, dayId, quotaKey, usageKey } from "./quota";

/** Browser-safe key row: never the hash. */
export type ApiKeyView = {
	id: string;
	name: string;
	prefix: string;
	createdAt: string;
	lastUsedAt: string | null;
	revokedAt: string | null;
};

export async function listApiKeys(db: Database, userId: string): Promise<ApiKeyView[]> {
	const rows = await db
		.select({
			id: apiKeys.id,
			name: apiKeys.name,
			prefix: apiKeys.prefix,
			createdAt: apiKeys.createdAt,
			lastUsedAt: apiKeys.lastUsedAt,
			revokedAt: apiKeys.revokedAt,
		})
		.from(apiKeys)
		.where(eq(apiKeys.userId, userId))
		.orderBy(desc(apiKeys.createdAt));
	return rows.map((r) => ({
		...r,
		createdAt: r.createdAt.toISOString(),
		lastUsedAt: r.lastUsedAt?.toISOString() ?? null,
		revokedAt: r.revokedAt?.toISOString() ?? null,
	}));
}

export type UsageDay = { day: string; count: number };

/**
 * Last `days` days, oldest first, zero-filled. Closed days come from `api_usage_daily` (flushed daily by cron);
 * today comes live from the quota store.
 */
export async function getUsageSeries(
	db: Database,
	store: CounterStore | null,
	userId: string,
	now = new Date(),
	days = 30,
): Promise<UsageDay[]> {
	const series = Array.from({ length: days }, (_, i) => dayId(new Date(now.getTime() - (days - 1 - i) * 86_400_000)));
	const first = series[0] as string;
	const today = dayId(now);
	const [rows, live] = await Promise.all([
		db
			.select({ day: apiUsageDaily.day, count: apiUsageDaily.count })
			.from(apiUsageDaily)
			.where(and(eq(apiUsageDaily.userId, userId), gte(apiUsageDaily.day, first))),
		store ? store.hget(usageKey(today), userId) : Promise.resolve(null),
	]);
	const byDay = new Map(rows.map((r) => [r.day, r.count]));
	if (live !== null) byDay.set(today, Math.max(live, byDay.get(today) ?? 0));
	return series.map((day) => ({ day, count: byDay.get(day) ?? 0 }));
}

export async function getQuotaUsed(store: CounterStore | null, userId: string, plan: PlanId): Promise<number> {
	const used = store ? await store.get(quotaKey(userId, plan)) : null;
	return Math.min(Math.max(used ?? 0, 0), PLAN_LIMITS[plan].quota);
}
