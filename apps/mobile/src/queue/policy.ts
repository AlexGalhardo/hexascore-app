import { QUEUE_MAX_AGE_MS, type QueueEntry } from "./store";

/** status 0 is a network failure (no response). */
export type SendOutcome =
	| { ok: true; body: unknown }
	| { ok: false; status: number; code: string | null; retryAfter: number | null };

export type Decision =
	| { kind: "applied"; body: unknown }
	| { kind: "retry" }
	| { kind: "rollback"; code: string | null; retryAfter: number | null }
	| { kind: "drop"; toast: boolean }
	| { kind: "pause" }
	| { kind: "halt" };

export const BACKOFF_MS = [2000, 10_000, 60_000] as const;
export const MAX_SERVER_ATTEMPTS = 5;

const CONFLICT_CODES = new Set(["rating-not-released", "status-not-completed"]);

export function backoffFor(attempts: number): number {
	return BACKOFF_MS[Math.min(Math.max(attempts, 1), BACKOFF_MS.length) - 1] ?? 60_000;
}

export function isExpired(entry: QueueEntry, now: number): boolean {
	return now - entry.createdAt > QUEUE_MAX_AGE_MS;
}

// SDD 10 §7.3 response table; the server wins on conflict.
export function decide(outcome: SendOutcome, entry: QueueEntry, now: number): Decision {
	if (outcome.ok) return { kind: "applied", body: outcome.body };
	const { status, code, retryAfter } = outcome;
	if (status === 426) return { kind: "pause" };
	// The session wipes the token; the queue stays bound to the user and replays after re-login.
	if (status === 401) return { kind: "halt" };
	if (status === 404) return { kind: "drop", toast: false };
	// 409 without a domain code is vukem's "key in flight": the first request is still running.
	if (status === 409 && (code === null || !CONFLICT_CODES.has(code))) {
		return isExpired(entry, now) ? { kind: "drop", toast: true } : { kind: "retry" };
	}
	if (status >= 400 && status < 500) return { kind: "rollback", code, retryAfter };
	if (isExpired(entry, now)) return { kind: "drop", toast: true };
	if (status >= 500 && entry.attempts + 1 >= MAX_SERVER_ATTEMPTS)
		return { kind: "drop", toast: true };
	return { kind: "retry" };
}
