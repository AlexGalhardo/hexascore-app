import { sql } from "drizzle-orm";
import type { Database } from "./db/client";
import { rateLimit } from "./db/schema";

// Fixed window on the Better Auth `rate_limit` table (keys are prefixed, so they never collide with its own).
// One atomic upsert per hit; good enough for low-volume forms. The public API uses Upstash instead.
export async function hitRateLimit(db: Database, key: string, max: number, windowMs: number): Promise<boolean> {
	const now = Date.now();
	const expired = sql`${rateLimit.lastRequest} < ${now - windowMs}`;
	const [row] = await db
		.insert(rateLimit)
		.values({ id: crypto.randomUUID(), key, count: 1, lastRequest: now })
		.onConflictDoUpdate({
			target: rateLimit.key,
			set: {
				count: sql`case when ${expired} then 1 else ${rateLimit.count} + 1 end`,
				lastRequest: sql`case when ${expired} then ${now} else ${rateLimit.lastRequest} end`,
			},
		})
		.returning({ count: rateLimit.count });
	return (row?.count ?? 1) <= max;
}

export function clientIp(headers: Headers): string {
	return headers.get("x-forwarded-for")?.split(",")[0]?.trim() || headers.get("x-real-ip") || "unknown";
}
