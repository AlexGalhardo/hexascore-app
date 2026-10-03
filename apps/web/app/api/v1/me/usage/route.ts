import { apiRoute } from "@/lib/api/handler";
import { creditsCacheKey, quotaKey } from "@/lib/api/quota";
import { emptyQuerySchema, usageSchema } from "@/lib/api/schemas";
import { creditBalance } from "@/lib/billing/subscriptions";
import { db } from "@/lib/db/client";

export const GET = apiRoute(
	{ query: emptyQuerySchema, response: usageSchema },
	async ({ userId, plan, limits, store }) => {
		const now = new Date();
		const [used, cachedCredits] = await Promise.all([
			store.get(quotaKey(userId, plan, now)),
			store.get(creditsCacheKey(userId)),
		]);
		// Redis holds the live balance once a paid request touched it; otherwise the ledger is the truth.
		const credits = cachedCredits ?? (await creditBalance(db, userId));
		const usedClamped = Math.min(Math.max(used ?? 0, 0), limits.quota);
		return {
			plan,
			quota: {
				limit: limits.quota,
				used: usedClamped,
				remaining: limits.quota - usedClamped,
				period: limits.period,
				resets_at:
					limits.period === "month"
						? new Date(Date.UTC(now.getUTCFullYear(), now.getUTCMonth() + 1, 1)).toISOString()
						: null,
			},
			credits: { balance: Math.max(credits, 0) },
			rate_limit: { requests_per_second: limits.requestsPerSecond },
		};
	},
);
