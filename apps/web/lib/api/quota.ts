import { Ratelimit } from "@upstash/ratelimit";
import { Redis } from "@upstash/redis";
import { sql } from "drizzle-orm";
import { db } from "../db/client";
import { env } from "../env";
import { PLAN_LIMITS, type PlanId } from "./plans";

export type QuotaDecision =
	| { ok: true; limit: number; remaining: number; resetSeconds: number; usedCredit: boolean }
	| { ok: false; reason: "rate_limited"; retryAfterSeconds: number }
	| { ok: false; reason: "quota_exhausted"; limit: number; resetSeconds: number };

/** Minimal counter surface shared by Upstash and the local in-memory fallback. */
export type CounterStore = {
	incr(key: string, ttlSeconds?: number): Promise<number>;
	decr(key: string): Promise<number>;
	get(key: string): Promise<number | null>;
	setIfAbsent(key: string, value: number): Promise<void>;
	/** Returns seconds to wait when over the per-second limit, 0 when allowed. */
	throttle(key: string, perSecond: number): Promise<number>;
	hincr(key: string, field: string, by: number): Promise<void>;
	hgetall(key: string): Promise<Record<string, number>>;
	hget(key: string, field: string): Promise<number | null>;
	del(key: string): Promise<void>;
	/** Atomic INCRBY that leaves an absent key absent; returns the new value or null. */
	incrByIfExists(key: string, by: number): Promise<number | null>;
};

function redisStore(redis: Redis): CounterStore {
	const limiters = new Map<number, Ratelimit>();
	return {
		async incr(key, ttl) {
			const value = await redis.incr(key);
			if (ttl && value === 1) await redis.expire(key, ttl);
			return value;
		},
		decr: (key) => redis.decr(key),
		get: async (key) => {
			const v = await redis.get<number | string>(key);
			return v === null ? null : Number(v);
		},
		async setIfAbsent(key, value) {
			await redis.set(key, value, { nx: true });
		},
		async hincr(key, field, by) {
			await redis.hincrby(key, field, by);
			await redis.expire(key, 7 * 86_400);
		},
		async hgetall(key) {
			const raw = (await redis.hgetall<Record<string, number | string>>(key)) ?? {};
			return Object.fromEntries(Object.entries(raw).map(([k, v]) => [k, Number(v)]));
		},
		async hget(key, field) {
			const v = await redis.hget<number | string>(key, field);
			return v === null ? null : Number(v);
		},
		async del(key) {
			await redis.del(key);
		},
		async incrByIfExists(key, by) {
			const v = await redis.eval<[number], number | null>(
				"if redis.call('exists', KEYS[1]) == 1 then return redis.call('incrby', KEYS[1], ARGV[1]) end return nil",
				[key],
				[by],
			);
			return v === null ? null : Number(v);
		},
		async throttle(key, perSecond) {
			let limiter = limiters.get(perSecond);
			if (!limiter) {
				limiter = new Ratelimit({ redis, limiter: Ratelimit.slidingWindow(perSecond, "1 s"), prefix: "hx:rl" });
				limiters.set(perSecond, limiter);
			}
			const result = await limiter.limit(key);
			return result.success ? 0 : Math.max(1, Math.ceil((result.reset - Date.now()) / 1000));
		},
	};
}

// ponytail: per-process memory, only for local dev/tests; production requires Upstash (see createStore).
export function memoryStore(): CounterStore {
	const values = new Map<string, number>();
	const hashes = new Map<string, Map<string, number>>();
	const windows = new Map<string, { second: number; count: number }>();
	return {
		async incr(key) {
			const v = (values.get(key) ?? 0) + 1;
			values.set(key, v);
			return v;
		},
		async decr(key) {
			const v = (values.get(key) ?? 0) - 1;
			values.set(key, v);
			return v;
		},
		get: async (key) => values.get(key) ?? null,
		async setIfAbsent(key, value) {
			if (!values.has(key)) values.set(key, value);
		},
		async hincr(key, field, by) {
			const h = hashes.get(key) ?? new Map<string, number>();
			h.set(field, (h.get(field) ?? 0) + by);
			hashes.set(key, h);
		},
		hgetall: async (key) => Object.fromEntries(hashes.get(key) ?? []),
		hget: async (key, field) => hashes.get(key)?.get(field) ?? null,
		async del(key) {
			values.delete(key);
			hashes.delete(key);
		},
		async incrByIfExists(key, by) {
			const current = values.get(key);
			if (current === undefined) return null;
			values.set(key, current + by);
			return current + by;
		},
		async throttle(key, perSecond) {
			const second = Math.floor(Date.now() / 1000);
			const w = windows.get(key);
			if (!w || w.second !== second) {
				windows.set(key, { second, count: 1 });
				return 0;
			}
			w.count += 1;
			return w.count > perSecond ? 1 : 0;
		},
	};
}

function createStore(): CounterStore | null {
	if (env.UPSTASH_REDIS_REST_URL && env.UPSTASH_REDIS_REST_TOKEN) {
		return redisStore(new Redis({ url: env.UPSTASH_REDIS_REST_URL, token: env.UPSTASH_REDIS_REST_TOKEN }));
	}
	if (env.NODE_ENV === "production") return null;
	return memoryStore();
}

let store: CounterStore | null | undefined;
export function getStore(): CounterStore | null {
	if (store === undefined) store = createStore();
	return store;
}

function monthWindow(now: Date): { id: string; resetSeconds: number } {
	const next = Date.UTC(now.getUTCFullYear(), now.getUTCMonth() + 1, 1);
	return {
		id: `${now.getUTCFullYear()}${String(now.getUTCMonth() + 1).padStart(2, "0")}`,
		resetSeconds: Math.ceil((next - now.getTime()) / 1000),
	};
}

export function dayId(now = new Date()): string {
	return now.toISOString().slice(0, 10);
}

export const creditsCacheKey = (userId: string): string => `hx:credits:${userId}`;
export const usageKey = (day: string): string => `hx:usage:${day}`;
export const creditsUsedKey = (day: string): string => `hx:credits-used:${day}`;

export function quotaKey(userId: string, plan: PlanId, now = new Date()): string {
	return PLAN_LIMITS[plan].period === "lifetime" ? `hx:q:${userId}:life` : `hx:q:${userId}:${monthWindow(now).id}`;
}

async function creditBalanceFromLedger(userId: string): Promise<number> {
	const rows = await db.execute<{ balance: string | null }>(
		sql`select coalesce(sum(delta), 0) as balance from credit_ledger where user_id = ${userId}`,
	);
	return Number((rows as unknown as { balance: string | null }[])[0]?.balance ?? 0);
}

/**
 * Consumes one request: per-second throttle, then plan quota, then (paid plans only) prepaid credits.
 * ponytail: credits live in Redis and are debited to the ledger by the daily flush; if the key is lost,
 * it is rebuilt from the ledger and unflushed debits of that day are forgiven.
 */
export async function consumeRequest(
	s: CounterStore,
	userId: string,
	plan: PlanId,
	loadCredits: (userId: string) => Promise<number> = creditBalanceFromLedger,
	now = new Date(),
): Promise<QuotaDecision> {
	const limits = PLAN_LIMITS[plan];
	const wait = await s.throttle(`${userId}`, limits.requestsPerSecond);
	if (wait > 0) return { ok: false, reason: "rate_limited", retryAfterSeconds: wait };

	const window = monthWindow(now);
	const resetSeconds = limits.period === "lifetime" ? 0 : window.resetSeconds;
	const key = quotaKey(userId, plan, now);
	const used = await s.incr(key, limits.period === "month" ? resetSeconds + 86_400 : undefined);
	if (used <= limits.quota) {
		await s.hincr(usageKey(dayId(now)), userId, 1);
		return { ok: true, limit: limits.quota, remaining: limits.quota - used, resetSeconds, usedCredit: false };
	}
	await s.decr(key);

	if (plan !== "free") {
		const creditsKey = creditsCacheKey(userId);
		if ((await s.get(creditsKey)) === null) await s.setIfAbsent(creditsKey, await loadCredits(userId));
		const left = await s.decr(creditsKey);
		if (left >= 0) {
			await s.hincr(usageKey(dayId(now)), userId, 1);
			await s.hincr(creditsUsedKey(dayId(now)), userId, 1);
			return { ok: true, limit: limits.quota, remaining: 0, resetSeconds, usedCredit: true };
		}
		await s.incr(creditsKey);
	}
	return { ok: false, reason: "quota_exhausted", limit: limits.quota, resetSeconds };
}

/** Gives back a request that ended in a 5xx, so server failures never cost quota. */
export async function refundRequest(s: CounterStore, userId: string, plan: PlanId, usedCredit: boolean): Promise<void> {
	const day = dayId();
	await s.hincr(usageKey(day), userId, -1);
	if (usedCredit) {
		await s.incr(creditsCacheKey(userId));
		await s.hincr(creditsUsedKey(day), userId, -1);
	} else {
		await s.decr(quotaKey(userId, plan));
	}
}

/**
 * Mirrors a purchase into the cached balance. Call only after the ledger row is committed: an absent key is
 * left alone because the next request rebuilds it from the ledger (which already includes the purchase).
 */
export async function addCachedCredits(s: CounterStore, userId: string, amount: number): Promise<void> {
	await s.incrByIfExists(creditsCacheKey(userId), amount);
}
