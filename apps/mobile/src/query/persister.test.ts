import { onlineManager, QueryClient } from "@tanstack/react-query";
import {
	persistQueryClientRestore,
	persistQueryClientSave,
} from "@tanstack/react-query-persist-client";
import { ApiError } from "../../../../packages/app-client/src/api";
import { setNetworkState } from "../platform/network";
import { createQueryClient, wireQueryManagers } from "./client";
import {
	CACHE_KEY,
	cacheBuster,
	createSqlitePersister,
	dehydrateOptions,
	type KeyValueStorage,
	MAX_AGE_MS,
	MAX_BYTES,
	readFormulaVersion,
	trackFormulaVersion,
} from "./persister";

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

const buster = cacheBuster("1.6.0", 1);

// No gc timers, so Jest can exit after the run.
const newClient = (): QueryClient =>
	new QueryClient({
		defaultOptions: {
			queries: { gcTime: Number.POSITIVE_INFINITY },
			mutations: { gcTime: Number.POSITIVE_INFINITY },
		},
	});

async function save(client: QueryClient, storage: KeyValueStorage, b = buster): Promise<void> {
	await persistQueryClientSave({
		queryClient: client,
		persister: createSqlitePersister(storage),
		buster: b,
		dehydrateOptions,
	});
}

async function restore(storage: KeyValueStorage, b = buster): Promise<QueryClient> {
	const client = newClient();
	await persistQueryClientRestore({
		queryClient: client,
		persister: createSqlitePersister(storage),
		buster: b,
		maxAge: MAX_AGE_MS,
	});
	return client;
}

test("read queries on the allowlist are persisted and restored", async () => {
	const storage = memoryStorage();
	const client = newClient();
	client.setQueryData(["getItem", "42"], { id: "42", title: "Arrival" });
	await save(client, storage);
	const restored = await restore(storage);
	expect(restored.getQueryData(["getItem", "42"])).toEqual({ id: "42", title: "Arrival" });
});

test("a mutation result or a problem body is never persisted", async () => {
	const storage = memoryStorage();
	const client = newClient();
	await client
		.getMutationCache()
		.build(client, { mutationFn: async () => ({ token: "hx_app_secret" }) })
		.execute(undefined);
	const problem = new ApiError(403, { title: "Forbidden", code: "account-banned" } as never);
	await client
		.fetchQuery({
			queryKey: ["getItem", "7"],
			queryFn: () => Promise.reject(problem),
			retry: false,
		})
		.catch(() => undefined);
	client.setQueryData(["tokensIndex"], { data: [{ id: "t1" }] });
	client.setQueryData(["somethingElse"], { secret: true });
	await save(client, storage);

	const raw = storage.data.get(CACHE_KEY) ?? "";
	expect(raw).not.toContain("hx_app_secret");
	expect(raw).not.toContain("account-banned");
	const stored = JSON.parse(raw) as { clientState: { mutations: unknown[]; queries: unknown[] } };
	expect(stored.clientState.mutations).toHaveLength(0);
	expect(stored.clientState.queries).toHaveLength(0);
});

test("data older than 24 h is discarded", async () => {
	const storage = memoryStorage();
	const client = newClient();
	client.setQueryData(["getItem", "1"], { id: "1" });
	await save(client, storage);
	const stored = JSON.parse(storage.data.get(CACHE_KEY) ?? "") as { timestamp: number };
	stored.timestamp = Date.now() - MAX_AGE_MS - 60_000;
	storage.data.set(CACHE_KEY, JSON.stringify(stored));

	const restored = await restore(storage);
	expect(restored.getQueryData(["getItem", "1"])).toBeUndefined();
	expect(storage.data.has(CACHE_KEY)).toBe(false);
});

test("a formulaVersion change busts the cache", async () => {
	const storage = memoryStorage();
	const client = newClient();
	client.setQueryData(["getItem", "1"], { id: "1", score: "8.1" });
	await save(client, storage, cacheBuster("1.6.0", 1));
	const restored = await restore(storage, cacheBuster("1.6.0", 2));
	expect(restored.getQueryData(["getItem", "1"])).toBeUndefined();
	expect(cacheBuster("1.6.0", 1)).not.toBe(cacheBuster("1.6.1", 1));
});

test("the last seen formulaVersion is kept for the next launch's buster", async () => {
	const storage = memoryStorage();
	const client = newClient();
	expect(readFormulaVersion(storage)).toBeNull();
	const stop = trackFormulaVersion(client, storage);
	await client.fetchQuery({
		queryKey: ["getItem", "1"],
		queryFn: async () => ({ id: "1", formulaVersion: 2 }),
	});
	stop();
	expect(readFormulaVersion(storage)).toBe(2);
});

test("the persisted size stays at or under 10 MB with a 2,000-item fixture", async () => {
	const storage = memoryStorage();
	const client = newClient();
	// About 6 KB per item: 2,000 of them are ~12 MB, over the budget.
	const blurb = "x".repeat(6_000);
	for (let i = 0; i < 2_000; i++) {
		client.setQueryData(
			["getItem", String(i)],
			{ id: String(i), description: blurb },
			{
				updatedAt: 1_000 + i,
			},
		);
	}
	await save(client, storage);
	const raw = storage.data.get(CACHE_KEY) ?? "";
	expect(new TextEncoder().encode(raw).length).toBeLessThanOrEqual(MAX_BYTES);
	const restored = await restore(storage);
	// The newest entries survive the trim, the oldest go first.
	expect(restored.getQueryData(["getItem", "1999"])).toBeDefined();
	expect(restored.getQueryData(["getItem", "0"])).toBeUndefined();
});

test("query defaults follow the legacy app and the platform drives online state", () => {
	const client = createQueryClient();
	expect(client.getDefaultOptions().queries).toMatchObject({ staleTime: 30_000, retry: 1 });
	const unwire = wireQueryManagers();
	setNetworkState("offline");
	expect(onlineManager.isOnline()).toBe(false);
	setNetworkState("online");
	expect(onlineManager.isOnline()).toBe(true);
	unwire();
	setNetworkState("unknown");
});
