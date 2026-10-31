import type { IdempotencyKey } from "../api/client";
import type { KeyValueStorage } from "../query/persister";

/** One queue lane per item and resource; `list:<listId>` is a single list add/remove (SDD 10 §7.3). */
export type QueueOp = "rating" | "status" | "timeToBeat" | `list:${string}`;
export type QueuePayload = { method: "put"; body?: unknown } | { method: "delete" };

export type QueueEntry = {
	id: string;
	op: QueueOp;
	itemId: string;
	payload: QueuePayload;
	idempotencyKey: IdempotencyKey;
	createdAt: number;
	attempts: number;
};

export type QueuedWrite = Pick<QueueEntry, "op" | "itemId" | "payload">;

export type MutationStore = {
	/** The user the queued writes belong to; they replay only for that user. */
	owner(): string | null;
	entries(): readonly QueueEntry[];
	/** Returns null when the queue is full (50 entries). */
	enqueue(userId: string, write: QueuedWrite): Promise<QueueEntry | null>;
	update(entry: QueueEntry): Promise<void>;
	remove(id: string): Promise<void>;
	clear(): Promise<void>;
};

export const QUEUE_KEY = "hexascore.mutationQueue";
export const MAX_ENTRIES = 50;
export const QUEUE_MAX_AGE_MS = 72 * 60 * 60 * 1000;

type Persisted = { userId: string | null; entries: QueueEntry[] };

function load(storage: KeyValueStorage): Persisted {
	const raw = storage.getItemSync(QUEUE_KEY);
	if (raw === null) return { userId: null, entries: [] };
	try {
		const parsed = JSON.parse(raw) as Partial<Persisted>;
		return {
			userId: typeof parsed.userId === "string" ? parsed.userId : null,
			entries: Array.isArray(parsed.entries) ? parsed.entries : [],
		};
	} catch {
		return { userId: null, entries: [] };
	}
}

export function createMutationStore(
	storage: KeyValueStorage,
	newKey: () => IdempotencyKey,
	now: () => number = Date.now,
): MutationStore {
	let state = load(storage);

	const save = (next: Persisted): Promise<void> => {
		state = next;
		return storage.setItem(QUEUE_KEY, JSON.stringify(next));
	};

	return {
		owner: () => state.userId,
		entries: () => state.entries,
		async enqueue(userId, write) {
			const kept = (state.userId === userId ? state.entries : []).filter(
				(e) => !(e.itemId === write.itemId && e.op === write.op),
			);
			if (kept.length >= MAX_ENTRIES) return null;
			// A fresh key per write: reusing one with a new body is a 422 (vukem idempotency).
			const key = newKey();
			const entry: QueueEntry = {
				...write,
				id: key,
				idempotencyKey: key,
				createdAt: now(),
				attempts: 0,
			};
			await save({ userId, entries: [...kept, entry] });
			return entry;
		},
		update: (entry) =>
			save({ ...state, entries: state.entries.map((e) => (e.id === entry.id ? entry : e)) }),
		remove: (id) => save({ ...state, entries: state.entries.filter((e) => e.id !== id) }),
		async clear() {
			state = { userId: null, entries: [] };
			await storage.removeItem(QUEUE_KEY);
		},
	};
}
