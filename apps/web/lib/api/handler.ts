import { and, eq, gt, isNull } from "drizzle-orm";
import type { z } from "zod";
import { db } from "../db/client";
import { apiKeys, subscriptions, user } from "../db/schema";
import { hashApiKey, parseBearer } from "./keys";
import { PLAN_LIMITS, type PlanId, type PlanLimits } from "./plans";
import { problem } from "./problem";
import { consumeRequest, getStore, refundRequest } from "./quota";

export type ApiContext = {
	requestId: string;
	userId: string;
	plan: PlanId;
	limits: PlanLimits;
	url: URL;
};

type KeyOwner = { userId: string; plan: PlanId; banned: boolean };

async function resolveKey(rawKey: string): Promise<KeyOwner | null> {
	const rows = await db
		.select({ userId: apiKeys.userId, banned: user.banned, plan: subscriptions.plan })
		.from(apiKeys)
		.innerJoin(user, eq(user.id, apiKeys.userId))
		.leftJoin(
			subscriptions,
			and(
				eq(subscriptions.userId, apiKeys.userId),
				eq(subscriptions.status, "active"),
				gt(subscriptions.currentPeriodEnd, new Date()),
			),
		)
		.where(and(eq(apiKeys.hash, hashApiKey(rawKey)), isNull(apiKeys.revokedAt)))
		.limit(1);
	const row = rows[0];
	if (!row) return null;
	return { userId: row.userId, plan: row.plan ?? "free", banned: row.banned ?? false };
}

function zodErrors(error: z.ZodError): { path: string; message: string }[] {
	return error.issues.map((i) => ({ path: i.path.join("."), message: i.message }));
}

type Handler<Q extends z.ZodType, P> = (ctx: ApiContext, query: z.infer<Q>, params: P) => Promise<unknown>;

/** Thrown by handlers for expected 404s; anything else becomes a 500 with a request id. */
export class NotFoundError extends Error {}

/**
 * Wraps a public API route: request id, API key auth, plan + quota, zod query validation,
 * RFC 9457 errors and rate-limit headers. 5xx responses are refunded to the caller's quota.
 */
export function apiRoute<Q extends z.ZodType, P = Record<string, never>>(
	querySchema: Q | ((limits: PlanLimits) => Q),
	handler: Handler<Q, P>,
) {
	return async (request: Request, context: { params: Promise<P> }): Promise<Response> => {
		const requestId = request.headers.get("x-request-id") ?? crypto.randomUUID();
		const key = parseBearer(request.headers.get("authorization"));
		if (!key) {
			return problem(
				"unauthorized",
				"Send your API key as `Authorization: Bearer hx_live_...`. Create one at /profile/api.",
				requestId,
				{},
				{ "www-authenticate": 'Bearer realm="api"' },
			);
		}

		const store = getStore();
		if (!store) {
			console.error(`[api] ${requestId} rate-limit store unavailable (UPSTASH_REDIS_REST_* missing)`);
			return problem(
				"service_unavailable",
				"The API is temporarily unavailable. Try again later.",
				requestId,
				{},
				{ "retry-after": "60" },
			);
		}

		let owner: KeyOwner | null;
		try {
			owner = await resolveKey(key);
		} catch (error) {
			console.error(`[api] ${requestId} key lookup failed`, error);
			return problem("internal_error", "Could not verify the API key. Try again later.", requestId);
		}
		if (!owner) return problem("unauthorized", "This API key does not exist or was revoked.", requestId);
		if (owner.banned) return problem("forbidden", "This account is suspended.", requestId);

		const limits = PLAN_LIMITS[owner.plan];
		const url = new URL(request.url);
		const schema = typeof querySchema === "function" ? querySchema(limits) : querySchema;
		const parsed = schema.safeParse(Object.fromEntries(url.searchParams));
		if (!parsed.success) {
			return problem("bad_request", "One or more query parameters are invalid.", requestId, {
				errors: zodErrors(parsed.error),
			});
		}

		const decision = await consumeRequest(store, owner.userId, owner.plan);
		if (!decision.ok && decision.reason === "rate_limited") {
			return problem(
				"rate_limited",
				`Plan '${owner.plan}' allows ${limits.requestsPerSecond} request(s) per second.`,
				requestId,
				{},
				{ "retry-after": String(decision.retryAfterSeconds) },
			);
		}
		if (!decision.ok) {
			return problem(
				"payment_required",
				owner.plan === "free"
					? "The Free plan includes 100 requests in total. Upgrade at /pricing to keep going."
					: "Monthly quota and credits are exhausted. Buy credits or upgrade at /pricing.",
				requestId,
				{ plan: owner.plan, limit: decision.limit },
			);
		}

		const headers = {
			"x-request-id": requestId,
			"ratelimit-limit": String(decision.limit),
			"ratelimit-remaining": String(decision.remaining),
			"ratelimit-reset": String(decision.resetSeconds),
			"x-quota-remaining": String(decision.remaining),
			"cache-control": "private, no-store",
		};

		try {
			const body = await handler(
				{ requestId, userId: owner.userId, plan: owner.plan, limits, url },
				parsed.data,
				await context.params,
			);
			return Response.json(body, { headers });
		} catch (error) {
			if (error instanceof NotFoundError) return problem("not_found", error.message, requestId, {}, headers);
			console.error(`[api] ${requestId} ${request.method} ${url.pathname} failed`, error);
			await refundRequest(store, owner.userId, owner.plan, decision.usedCredit).catch(() => undefined);
			return problem("internal_error", "Unexpected error. It was not counted against your quota.", requestId);
		}
	};
}
