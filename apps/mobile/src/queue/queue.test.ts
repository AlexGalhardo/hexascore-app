import { ApiError } from "../../../../packages/app-client/src/api";
import type { AppClient, IdempotencyKey } from "../api/client";
import type { KeyValueStorage } from "../query/persister";
import { createDrain, type QueueResult, sendEntry, startQueue } from "./drain";
import { BACKOFF_MS, decide, type SendOutcome } from "./policy";
import {
	createMutationStore,
	MAX_ENTRIES,
	type MutationStore,
	QUEUE_KEY,
	QUEUE_MAX_AGE_MS,
	type QueueEntry,
} from "./store";

function memoryStorage(): KeyValueStorage & { data: Map<string, string> } {
	const data = new Map<string, string>();
	return {
		data,
		getItem: async (key) => data.get(key) ?? null,
		getItemSync: (key) => data.get(key) ?? null,
		setItem: async (key, value) => {
			data.set(key, value);
		},
		removeItem: async (key) => {
			data.delete(key);
		},
	};
}

let keyCounter = 0;
const nextKey = (): IdempotencyKey => `key-${++keyCounter}` as IdempotencyKey;

const T0 = 1_800_000_000_000;

function setup(start: number = T0) {
	const storage = memoryStorage();
	let now = start;
	const clock = { now: () => now, advance: (ms: number) => (now += ms) };
	const store = createMutationStore(storage, nextKey, clock.now);
	return { storage, store, clock };
}

const rating = (itemId: string, score: number) => ({
	op: "rating" as const,
	itemId,
	payload: { method: "put" as const, body: { scores: { gameplay: score } } },
});

const fail = (
	status: number,
	code: string | null = null,
	retryAfter: number | null = null,
): SendOutcome => ({
	ok: false,
	status,
	code,
	retryAfter,
});

function entry(overrides: Partial<QueueEntry> = {}): QueueEntry {
	return {
		id: "e1",
		op: "rating",
		itemId: "i1",
		payload: { method: "delete" },
		idempotencyKey: "k" as IdempotencyKey,
		createdAt: T0,
		attempts: 0,
		...overrides,
	};
}

describe("store", () => {
	test("persists entries with the documented shape and reloads them", async () => {
		const { storage, store, clock } = setup();
		await store.enqueue("u1", rating("i1", 7));
		const reloaded = createMutationStore(storage, nextKey, clock.now);
		const [first] = reloaded.entries();
		expect(Object.keys(first ?? {}).sort()).toEqual(
			["attempts", "createdAt", "id", "idempotencyKey", "itemId", "op", "payload"].sort(),
		);
		expect(reloaded.owner()).toBe("u1");
		expect(JSON.parse(storage.data.get(QUEUE_KEY) ?? "{}").userId).toBe("u1");
	});

	test("replace: a newer write for the same (itemId, op) replaces the queued one with a new key", async () => {
		const { store } = setup();
		const old = await store.enqueue("u1", rating("i1", 7));
		await store.enqueue("u1", rating("i2", 5));
		const fresh = await store.enqueue("u1", rating("i1", 9));
		const forItem = store.entries().filter((e) => e.itemId === "i1");
		expect(forItem).toHaveLength(1);
		expect(forItem[0]?.idempotencyKey).not.toBe(old?.idempotencyKey);
		expect(forItem[0]?.payload).toEqual(fresh?.payload);
		expect(store.entries().map((e) => e.itemId)).toEqual(["i2", "i1"]);
	});

	test("replace keeps other ops on the same item", async () => {
		const { store } = setup();
		await store.enqueue("u1", rating("i1", 7));
		await store.enqueue("u1", {
			op: "status",
			itemId: "i1",
			payload: { method: "put", body: { status: "playing" } },
		});
		expect(store.entries()).toHaveLength(2);
	});

	test("refuses a write past 50 entries", async () => {
		const { store } = setup();
		for (let i = 0; i < MAX_ENTRIES; i++) await store.enqueue("u1", rating(`i${i}`, 5));
		expect(await store.enqueue("u1", rating("extra", 5))).toBeNull();
		expect(store.entries()).toHaveLength(MAX_ENTRIES);
	});

	test("another user's enqueue discards the previous user's entries", async () => {
		const { store } = setup();
		await store.enqueue("u1", rating("i1", 7));
		await store.enqueue("u2", rating("i2", 7));
		expect(store.entries().map((e) => e.itemId)).toEqual(["i2"]);
		expect(store.owner()).toBe("u2");
	});

	test("an unparsable row starts an empty queue", () => {
		const storage = memoryStorage();
		storage.data.set(QUEUE_KEY, "{not json");
		expect(createMutationStore(storage, nextKey).entries()).toEqual([]);
	});

	test("clear removes the row", async () => {
		const { storage, store } = setup();
		await store.enqueue("u1", rating("i1", 7));
		await store.clear();
		expect(storage.data.has(QUEUE_KEY)).toBe(false);
		expect(store.entries()).toEqual([]);
	});
});

describe("policy (SDD 10 §7.3)", () => {
	test("2xx applies the server body", () => {
		expect(decide({ ok: true, body: { rating: 1 } }, entry(), T0)).toEqual({
			kind: "applied",
			body: { rating: 1 },
		});
	});

	test("409 in flight retries", () => {
		expect(decide(fail(409), entry(), T0).kind).toBe("retry");
	});

	test("409 rating-not-released and status-not-completed roll back", () => {
		expect(decide(fail(409, "rating-not-released"), entry(), T0)).toMatchObject({
			kind: "rollback",
			code: "rating-not-released",
		});
		expect(decide(fail(409, "status-not-completed"), entry(), T0).kind).toBe("rollback");
	});

	test("422 rolls back", () => {
		expect(decide(fail(422, "validation-failed"), entry(), T0).kind).toBe("rollback");
	});

	test("429 item-ratings-limited rolls back with Retry-After", () => {
		expect(decide(fail(429, "item-ratings-limited", 3600), entry(), T0)).toEqual({
			kind: "rollback",
			code: "item-ratings-limited",
			retryAfter: 3600,
		});
	});

	test("404 drops the entry without a toast", () => {
		expect(decide(fail(404), entry(), T0)).toEqual({ kind: "drop", toast: false });
	});

	test("5xx retries until the fifth attempt, then drops with a toast", () => {
		expect(decide(fail(503), entry({ attempts: 3 }), T0).kind).toBe("retry");
		expect(decide(fail(503), entry({ attempts: 4 }), T0)).toEqual({ kind: "drop", toast: true });
	});

	test("an entry older than 72 h drops with a toast", () => {
		expect(decide(fail(503), entry(), T0 + QUEUE_MAX_AGE_MS + 1)).toEqual({
			kind: "drop",
			toast: true,
		});
	});

	test("a network failure retries without a cap below 72 h", () => {
		expect(decide(fail(0), entry({ attempts: 20 }), T0).kind).toBe("retry");
	});

	test("426 pauses and 401 halts", () => {
		expect(decide(fail(426), entry(), T0).kind).toBe("pause");
		expect(decide(fail(401), entry(), T0).kind).toBe("halt");
	});

	test("backoff is 2 s, 10 s, then 60 s", () => {
		expect(BACKOFF_MS).toEqual([2000, 10_000, 60_000]);
	});
});

type Harness = ReturnType<typeof setup> & {
	store: MutationStore;
	sent: QueueEntry[];
	results: Array<{ entry: QueueEntry; result: QueueResult }>;
	timers: Array<{ fn: () => void; ms: number }>;
	state: { user: string | null; paused: boolean; online: boolean };
	outcomes: SendOutcome[];
	drain: () => Promise<void>;
};

function harness(outcomes: SendOutcome[] = []): Harness {
	const base = setup();
	const sent: QueueEntry[] = [];
	const results: Harness["results"] = [];
	const timers: Harness["timers"] = [];
	const state = { user: "u1" as string | null, paused: false, online: true };
	const drainer = createDrain({
		store: base.store,
		send: async (e) => {
			sent.push(e);
			return outcomes.shift() ?? { ok: true, body: null };
		},
		currentUser: () => state.user,
		paused: () => state.paused,
		online: () => state.online,
		onResult: (e, result) => results.push({ entry: e, result }),
		now: base.clock.now,
		schedule: (fn, ms) => {
			const timer = { fn, ms };
			timers.push(timer);
			return () => timers.splice(timers.indexOf(timer), 1);
		},
	});
	return { ...base, sent, results, timers, state, outcomes, drain: drainer.drain };
}

describe("drain", () => {
	test("drains FIFO and removes applied entries", async () => {
		const h = harness([
			{ ok: true, body: "a" },
			{ ok: true, body: "b" },
		]);
		await h.store.enqueue("u1", rating("i1", 7));
		await h.store.enqueue("u1", rating("i2", 7));
		await h.drain();
		expect(h.sent.map((e) => e.itemId)).toEqual(["i1", "i2"]);
		expect(h.results.map((r) => r.result)).toEqual([
			{ kind: "applied", body: "a" },
			{ kind: "applied", body: "b" },
		]);
		expect(h.store.entries()).toEqual([]);
	});

	test("409 in flight keeps the entry, counts the attempt and backs off 2 s, 10 s, 60 s", async () => {
		const h = harness([fail(409), fail(409), fail(409), fail(409)]);
		await h.store.enqueue("u1", rating("i1", 7));
		await h.drain();
		expect(h.store.entries()[0]?.attempts).toBe(1);
		expect(h.timers.map((t) => t.ms)).toEqual([2000]);
		for (const expected of [10_000, 60_000, 60_000]) {
			const timer = h.timers.shift();
			timer?.fn();
			await new Promise((r) => setImmediate(r));
			expect(h.timers.at(-1)?.ms).toBe(expected);
		}
		expect(h.sent.every((e) => e.idempotencyKey === h.sent[0]?.idempotencyKey)).toBe(true);
	});

	test("rollback and drop remove the entry and report it", async () => {
		const h = harness([
			fail(409, "rating-not-released"),
			fail(404),
			fail(429, "item-ratings-limited", 60),
		]);
		await h.store.enqueue("u1", rating("i1", 7));
		await h.store.enqueue("u1", rating("i2", 7));
		await h.store.enqueue("u1", rating("i3", 7));
		await h.drain();
		expect(h.results.map((r) => r.result.kind)).toEqual(["rollback", "drop", "rollback"]);
		expect(h.results[2]?.result).toEqual({
			kind: "rollback",
			code: "item-ratings-limited",
			retryAfter: 60,
		});
		expect(h.store.entries()).toEqual([]);
	});

	test("5xx after 5 attempts drops with a toast", async () => {
		const h = harness([fail(500), fail(500), fail(500), fail(500), fail(500)]);
		await h.store.enqueue("u1", rating("i1", 7));
		for (let i = 0; i < 5; i++) await h.drain();
		expect(h.sent).toHaveLength(5);
		expect(h.results).toEqual([expect.objectContaining({ result: { kind: "drop", toast: true } })]);
		expect(h.store.entries()).toEqual([]);
	});

	test("an entry older than 72 h is dropped with a toast and never sent", async () => {
		const h = harness();
		await h.store.enqueue("u1", rating("i1", 7));
		h.clock.advance(QUEUE_MAX_AGE_MS + 1);
		await h.drain();
		expect(h.sent).toEqual([]);
		expect(h.results[0]?.result).toEqual({ kind: "drop", toast: true });
	});

	test("pauses on 426 and does not drain while an update is required", async () => {
		const h = harness([fail(426)]);
		await h.store.enqueue("u1", rating("i1", 7));
		await h.drain();
		expect(h.store.entries()).toHaveLength(1);
		h.state.paused = true;
		await h.drain();
		expect(h.sent).toHaveLength(1);
		expect(h.timers).toEqual([]);
	});

	test("does nothing offline", async () => {
		const h = harness();
		await h.store.enqueue("u1", rating("i1", 7));
		h.state.online = false;
		await h.drain();
		expect(h.sent).toEqual([]);
	});

	test("replays only for the user who created the queue; another user discards it", async () => {
		const h = harness();
		await h.store.enqueue("u1", rating("i1", 7));
		h.state.user = null;
		await h.drain();
		expect(h.store.entries()).toHaveLength(1);
		h.state.user = "u2";
		await h.drain();
		expect(h.sent).toEqual([]);
		expect(h.store.entries()).toEqual([]);
	});

	test("a write replaced while in flight is not removed by the old result", async () => {
		const base = setup();
		let release: (o: SendOutcome) => void = () => {};
		const drainer = createDrain({
			store: base.store,
			send: () => new Promise<SendOutcome>((r) => (release = r)),
			currentUser: () => "u1",
			paused: () => false,
			online: () => true,
			onResult: () => {},
			now: base.clock.now,
			schedule: () => () => {},
		});
		await base.store.enqueue("u1", rating("i1", 7));
		const running = drainer.drain();
		await new Promise((r) => setImmediate(r));
		await base.store.enqueue("u1", rating("i1", 9));
		const second = drainer.drain();
		expect(second).toBe(running);
		release({ ok: true, body: null });
		await new Promise((r) => setImmediate(r));
		expect(base.store.entries().map((e) => e.payload)).toEqual([rating("i1", 9).payload]);
		release({ ok: true, body: null });
		await running;
		expect(base.store.entries()).toEqual([]);
	});
});

describe("startQueue", () => {
	test("drains on reconnect and on foreground, and stops on unsubscribe", () => {
		let network: (s: "online" | "offline" | "unknown") => void = () => {};
		let active: (a: boolean) => void = () => {};
		const drain = jest.fn(async () => {});
		const stop = startQueue(drain, {
			subscribeNetwork: (l) => {
				network = l;
				return () => (network = () => {});
			},
			subscribeAppActive: (l) => {
				active = l;
				return () => (active = () => {});
			},
		});
		expect(drain).toHaveBeenCalledTimes(1);
		network("offline");
		network("online");
		active(false);
		active(true);
		expect(drain).toHaveBeenCalledTimes(3);
		stop();
		network("online");
		expect(drain).toHaveBeenCalledTimes(3);
	});
});

describe("sendEntry", () => {
	function fakeClient(impl: (method: string, args: unknown) => Promise<unknown>) {
		const calls: Array<{ key: string; method: string; args: unknown }> = [];
		const client = new Proxy(
			{},
			{ get: (_t, method: string) => (args: unknown) => impl(method, args) },
		);
		const app = {
			write: async (key: IdempotencyKey, call: (c: never) => Promise<unknown>) => {
				const proxy = new Proxy(
					{},
					{
						get: (_t, method: string) => (args: unknown) => {
							calls.push({ key, method, args });
							return (client as Record<string, (a: unknown) => Promise<unknown>>)[method]?.(args);
						},
					},
				);
				return call(proxy as never);
			},
		} as unknown as AppClient;
		return { app, calls };
	}

	test.each([
		["rating", "put", "ratingsPut"],
		["rating", "delete", "ratingsDelete"],
		["status", "put", "statusPut"],
		["status", "delete", "statusDelete"],
		["timeToBeat", "put", "timeToBeatPut"],
		["timeToBeat", "delete", "timeToBeatDelete"],
	] as const)("%s %s calls %s with the entry key", async (op, method, expected) => {
		const { app, calls } = fakeClient(async () => ({ ok: 1 }));
		const payload = method === "put" ? { method, body: { x: 1 } } : { method };
		const outcome = await sendEntry(
			app,
			entry({ op, payload, idempotencyKey: "kk" as IdempotencyKey }),
		);
		expect(outcome).toEqual({ ok: true, body: { ok: 1 } });
		expect(calls[0]).toMatchObject({ key: "kk", method: expected, args: { params: { id: "i1" } } });
	});

	test("a list op adds or removes the item on that list", async () => {
		const { app, calls } = fakeClient(async () => null);
		await sendEntry(app, entry({ op: "list:L9", payload: { method: "put" } }));
		await sendEntry(app, entry({ op: "list:L9", payload: { method: "delete" } }));
		expect(calls.map((c) => [c.method, c.args])).toEqual([
			["listsItemsPut", { params: { id: "L9", itemId: "i1" } }],
			["listsItemsDelete", { params: { id: "L9", itemId: "i1" } }],
		]);
	});

	test("maps an ApiError to status, code and Retry-After, and a network error to status 0", async () => {
		const limited = Object.assign(
			new ApiError(429, { status: 429, code: "item-ratings-limited" } as never),
			{
				retryAfter: 120,
			},
		);
		const { app } = fakeClient(async () => {
			throw limited;
		});
		expect(await sendEntry(app, entry())).toEqual({
			ok: false,
			status: 429,
			code: "item-ratings-limited",
			retryAfter: 120,
		});
		const offline = fakeClient(async () => {
			throw new TypeError("Network request failed");
		});
		expect(await sendEntry(offline.app, entry())).toEqual({
			ok: false,
			status: 0,
			code: null,
			retryAfter: null,
		});
	});
});
