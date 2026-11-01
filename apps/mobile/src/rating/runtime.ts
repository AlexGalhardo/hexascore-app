import type { QueryClient } from "@tanstack/react-query";
import type { IdempotencyKey } from "../api/client";
import type { Session, SessionState } from "../auth/session";
import type { KeyValueStorage } from "../query/persister";
import { createDrain, type DrainDeps, type QueueResult } from "../queue/drain";
import type { SendOutcome } from "../queue/policy";
import { createMutationStore, type QueueEntry } from "../queue/store";
import { createStatusCache } from "../status/cache";
import { createRatingCache } from "./cache";
import type { RatingQueue } from "./useRateMutation";

export type RatingToast =
	| "ratings.notReleased"
	| "ratings.notSaved"
	| "item.status.notSaved"
	| "item.timeToBeat.notSaved"
	| "errors.status-not-completed";

export type RatingRuntimeDeps = {
	storage: KeyValueStorage;
	queryClient: QueryClient;
	newKey: () => IdempotencyKey;
	send: (entry: QueueEntry) => Promise<SendOutcome>;
	online: () => boolean;
	paused: () => boolean;
	notify: (message: RatingToast) => void;
	schedule?: DrainDeps["schedule"];
};

export type RatingRuntime = {
	queue: RatingQueue;
	/** The `/me` id the queued writes belong to; null while signed out. */
	setUser: (id: string | null) => void;
	drain: () => Promise<void>;
};

const NOT_SAVED: Partial<Record<QueueEntry["op"], RatingToast>> = {
	rating: "ratings.notSaved",
	status: "item.status.notSaved",
	timeToBeat: "item.timeToBeat.notSaved",
};

export function toastFor(result: QueueResult, op: QueueEntry["op"]): RatingToast | null {
	const notSaved = NOT_SAVED[op] ?? null;
	if (result.kind === "rollback") {
		if (result.code === "rating-not-released") return "ratings.notReleased";
		if (result.code === "status-not-completed") return "errors.status-not-completed";
		return notSaved;
	}
	return result.kind === "drop" && result.toast ? notSaved : null;
}

export function createRatingRuntime(deps: RatingRuntimeDeps): RatingRuntime {
	let user: string | null = null;
	const store = createMutationStore(deps.storage, deps.newKey);
	const cache = createRatingCache(deps.queryClient);
	const statusCache = createStatusCache(deps.queryClient);
	const { drain } = createDrain({
		store,
		send: deps.send,
		currentUser: () => user,
		online: deps.online,
		paused: deps.paused,
		schedule: deps.schedule,
		onResult(entry, result) {
			cache.onResult(entry, result);
			statusCache.onResult(entry, result);
			const toast = toastFor(result, entry.op);
			if (toast) deps.notify(toast);
		},
	});
	return {
		queue: {
			cache,
			statusCache,
			drain,
			enqueue: (write) => (user === null ? Promise.resolve(null) : store.enqueue(user, write)),
		},
		setUser(id) {
			user = id;
			if (id !== null) void drain();
		},
		drain,
	};
}

/** Reads `GET /me` on every active session: deletion routing, and the queue's user binding. */
export function followMe(
	session: Pick<Session, "subscribe" | "observeMe" | "handleError">,
	meShow: () => Promise<unknown>,
	setUser: (id: string | null) => void,
): () => void {
	return session.subscribe((state: SessionState) => {
		if (state.kind !== "active") {
			setUser(null);
			return;
		}
		meShow().then(
			(body) => {
				const me = (typeof body === "object" && body !== null ? body : {}) as {
					id?: unknown;
					deletionScheduledAt?: string | null;
				};
				session.observeMe(me);
				setUser(typeof me.id === "string" ? me.id : null);
			},
			(error: unknown) => void session.handleError(error),
		);
	});
}
