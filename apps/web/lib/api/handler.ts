import { and, eq, gt, isNull, lt, or } from "drizzle-orm";
import type { z } from "zod";
import { db } from "../db/client";
import { apiKeys, subscriptions, user } from "../db/schema";
import { hashApiKey, parseBearer } from "./keys";
import { PLAN_LIMITS, type PlanId, type PlanLimits } from "./plans";
import { problem } from "./problem";
import { type CounterStore, consumeRequest, getStore, refundRequest } from "./quota";

export type ApiContext = {
	requestId: string;
	userId: string;
	plan: PlanId;
	limits: PlanLimits;
	store: CounterStore;
	url: URL;
};

type KeyOwner = { keyId: string; userId: string; plan: PlanId; banned: boolean };

async function resolveKey(rawKey: string): Promise<KeyOwner | null> {
	const rows = await db
		.select({ keyId: apiKeys.id, userId: apiKeys.userId, banned: user.banned, plan: subscriptions.plan })
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
	return { keyId: row.keyId, userId: row.userId, plan: row.plan ?? "free", banned: row.banned ?? false };
}

// At most one write per key per minute, off the response path; a failure only delays the dashboard value.
function touchKey(keyId: string): void {
	const now = new Date();
	db.update(apiKeys)
		.set({ lastUsedAt: now })
		.where(
			and(
				eq(apiKeys.id, keyId),
				or(isNull(apiKeys.lastUsedAt), lt(apiKeys.lastUsedAt, new Date(now.getTime() - 60_000))),
			),
		)
		.catch((error: unknown) => console.warn("[api] last_used_at update failed", error));
}

function zodErrors(error: z.ZodError): { path: string; message: string }[] {
	return error.issues.map((i) => ({ path: i.path.join("."), message: i.message }));
}

/** Thrown by handlers for expected 404s; anything else becomes a 500 with a request id. */
export class NotFoundError extends Error {}

/** Thrown by handlers for input that only the handler can reject (e.g. a cursor from another query). */
export class BadRequestError extends Error {}

type RouteOptions<P, Q extends z.ZodType, R extends z.ZodType> = {
	/** Path params; a mismatch (unknown resource type, malformed id) is a 404. */
	params?: z.ZodType<P>;
	query: Q | ((limits: PlanLimits, params: P) => Q);
	/** Every body is parsed through this schema, so the OpenAPI contract and the wire format cannot drift. */
	response: R;
};

type Handler<P, Q extends z.ZodType, R extends z.ZodType> = (
	ctx: ApiContext,
	query: z.infer<Q>,
	params: P,
) => Promise<z.input<R>>;

/**
 * Wraps a public API route: request id, API key auth, plan + quota, zod query validation,
 * RFC 9457 errors and rate-limit headers. 5xx responses are refunded to the caller's quota.
 */
export function apiRoute<P, Q extends z.ZodType, R extends z.ZodType>(
	options: RouteOptions<P, Q, R>,
	handler: Handler<P, Q, R>,
): (request: Request, context: { params: Promise<unknown> }) => Promise<Response> {
	return async (request, context) => {
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

		const url = new URL(request.url);
		const rawParams = (await context.params) ?? {};
		const parsedParams = options.params ? options.params.safeParse(rawParams) : null;
		if (parsedParams && !parsedParams.success) {
			return problem("not_found", `No resource at ${url.pathname}.`, requestId);
		}
		const params = (parsedParams ? parsedParams.data : rawParams) as P;

		const limits = PLAN_LIMITS[owner.plan];
		const schema = typeof options.query === "function" ? options.query(limits, params) : options.query;
		// An empty value (`?min_score=`) is never meaningful, and z.coerce would silently read "" as 0.
		const empty = [...url.searchParams].filter(([, value]) => value.trim() === "").map(([key]) => key);
		if (empty.length > 0) {
			return problem("bad_request", "One or more query parameters are invalid.", requestId, {
				errors: empty.map((key) => ({ path: key, message: "Must not be empty." })),
			});
		}
		const parsed = schema.safeParse(Object.fromEntries(url.searchParams));
		if (!parsed.success) {
			return problem("bad_request", "One or more query parameters are invalid.", requestId, {
				errors: zodErrors(parsed.error),
			});
		}

		let decision: Awaited<ReturnType<typeof consumeRequest>>;
		try {
			decision = await consumeRequest(store, owner.userId, owner.plan);
		} catch (error) {
			// Fail closed: without the counter there is no quota enforcement, so no data is served.
			console.error(`[api] ${requestId} rate-limit store failed`, error);
			return problem(
				"service_unavailable",
				"The API is temporarily unavailable. Try again later.",
				requestId,
				{},
				{ "retry-after": "30" },
			);
		}
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
		touchKey(owner.keyId);

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
				{ requestId, userId: owner.userId, plan: owner.plan, limits, store, url },
				parsed.data,
				params,
			);
			return Response.json(options.response.parse(body), { headers });
		} catch (error) {
			if (error instanceof NotFoundError) return problem("not_found", error.message, requestId, {}, headers);
			// By name too: errors crossing module copies (tests, bundler chunks) fail `instanceof`.
			if (error instanceof BadRequestError || (error instanceof Error && error.name === "InvalidCursorError")) {
				const detail = error instanceof BadRequestError ? error.message : "The cursor is invalid or expired.";
				return problem("bad_request", detail, requestId, {}, headers);
			}
			console.error(`[api] ${requestId} ${request.method} ${url.pathname} failed`, error);
			await refundRequest(store, owner.userId, owner.plan, decision.usedCredit).catch(() => undefined);
			return problem("internal_error", "Unexpected error. It was not counted against your quota.", requestId);
		}
	};
}
