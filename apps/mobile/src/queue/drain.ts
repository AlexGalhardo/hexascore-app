import { ApiError } from "../../../../packages/app-client/src/api";
import type { AppClient } from "../api/client";
import type { NetworkState } from "../platform/network";
import { backoffFor, type Decision, decide, isExpired, type SendOutcome } from "./policy";
import type { MutationStore, QueueEntry } from "./store";

/** What the UI hears about a finished entry: replace the cache, roll back with a toast, or drop. */
export type QueueResult = Extract<Decision, { kind: "applied" | "rollback" | "drop" }>;

export type DrainDeps = {
	store: MutationStore;
	send: (entry: QueueEntry) => Promise<SendOutcome>;
	currentUser: () => string | null;
	/** True after a 426 (`appClient.upgradeRequired()`): the queue waits for the update. */
	paused: () => boolean;
	online: () => boolean;
	onResult: (entry: QueueEntry, result: QueueResult) => void;
	now?: () => number;
	schedule?: (fn: () => void, ms: number) => () => void;
};

const defaultSchedule = (fn: () => void, ms: number): (() => void) => {
	const id = setTimeout(fn, ms);
	return () => clearTimeout(id);
};

export function createDrain(deps: DrainDeps): { drain: () => Promise<void> } {
	const now = deps.now ?? Date.now;
	const schedule = deps.schedule ?? defaultSchedule;
	let running: Promise<void> | null = null;
	let cancelTimer: (() => void) | null = null;

	const settle = async (entry: QueueEntry, result: QueueResult): Promise<void> => {
		await deps.store.remove(entry.id);
		deps.onResult(entry, result);
	};

	const run = async (): Promise<void> => {
		while (!deps.paused() && deps.online()) {
			const user = deps.currentUser();
			if (user === null) return;
			if (deps.store.owner() !== null && deps.store.owner() !== user) {
				await deps.store.clear();
				return;
			}
			const entry = deps.store.entries()[0];
			if (entry === undefined) return;
			if (isExpired(entry, now())) {
				await settle(entry, { kind: "drop", toast: true });
				continue;
			}
			const decision = decide(await deps.send(entry), entry, now());
			switch (decision.kind) {
				case "pause":
				case "halt":
					return;
				case "retry": {
					const attempts = entry.attempts + 1;
					await deps.store.update({ ...entry, attempts });
					cancelTimer = schedule(() => void drain(), backoffFor(attempts));
					return;
				}
				default:
					await settle(entry, decision);
			}
		}
	};

	const drain = (): Promise<void> => {
		if (running) return running;
		cancelTimer?.();
		cancelTimer = null;
		running = run().finally(() => {
			running = null;
		});
		return running;
	};

	return { drain };
}

export type QueueTriggers = {
	subscribeNetwork: (listener: (state: NetworkState) => void) => () => void;
	subscribeAppActive: (listener: (active: boolean) => void) => () => void;
};

/** FIFO drain at start, on reconnect and on foreground (SDD 10 §7.3). */
export function startQueue(drain: () => Promise<void>, triggers: QueueTriggers): () => void {
	void drain();
	const offNetwork = triggers.subscribeNetwork((state) => {
		if (state === "online") void drain();
	});
	const offActive = triggers.subscribeAppActive((active) => {
		if (active) void drain();
	});
	return () => {
		offNetwork();
		offActive();
	};
}

function call(
	client: Parameters<Parameters<AppClient["write"]>[1]>[0],
	entry: QueueEntry,
): Promise<unknown> {
	const { itemId, payload } = entry;
	const params = { id: itemId };
	const body = payload.method === "put" ? payload.body : undefined;
	const put = payload.method === "put";
	if (entry.op.startsWith("list:")) {
		const list = { params: { id: entry.op.slice("list:".length), itemId } };
		return put ? client.listsItemsPut({ ...list, body }) : client.listsItemsDelete(list);
	}
	switch (entry.op) {
		case "rating":
			return put ? client.ratingsPut({ params, body }) : client.ratingsDelete({ params });
		case "status":
			return put ? client.statusPut({ params, body }) : client.statusDelete({ params });
		default:
			return put ? client.timeToBeatPut({ params, body }) : client.timeToBeatDelete({ params });
	}
}

export async function sendEntry(app: AppClient, entry: QueueEntry): Promise<SendOutcome> {
	try {
		return {
			ok: true,
			body: await app.write(entry.idempotencyKey, (client) => call(client, entry)),
		};
	} catch (error) {
		if (!(error instanceof ApiError)) return { ok: false, status: 0, code: null, retryAfter: null };
		// The generated ApiProblem omits RFC 9457 extensions (VF-NEW-P5.A.3.1); read `code` as unknown.
		const code: unknown = (error.problem as { code?: unknown } | undefined)?.code;
		const retryAfter: unknown = (error as { retryAfter?: unknown }).retryAfter;
		return {
			ok: false,
			status: error.status,
			code: typeof code === "string" ? code : null,
			retryAfter: typeof retryAfter === "number" ? retryAfter : null,
		};
	}
}
