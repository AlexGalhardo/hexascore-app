import type { DehydrateOptions, QueryClient } from "@tanstack/react-query";
import type { PersistedClient, Persister } from "@tanstack/react-query-persist-client";
import type { ReadClient } from "../api/client";

/** The subset of expo-sqlite/kv-store the cache needs; a Map stands in for it in tests. */
export type KeyValueStorage = {
	getItem(key: string): Promise<string | null>;
	getItemSync(key: string): string | null;
	setItem(key: string, value: string): Promise<void>;
	removeItem(key: string): Promise<void>;
};

export const CACHE_KEY = "hexascore.queryCache";
const FORMULA_KEY = "hexascore.formulaVersion";
export const MAX_BYTES = 10 * 1024 * 1024;
export const MAX_AGE_MS = 24 * 60 * 60 * 1000;

// Allowlist (ROADMAP P5.A.4.1): only catalogue and own-data reads; tokensIndex and anything new stay in memory.
const PERSISTED_QUERIES: ReadonlySet<string> = new Set<keyof ReadClient>([
	"getEntity",
	"getItem",
	"library",
	"listBookGenres",
	"listBookTrending",
	"listEditions",
	"listEntityItems",
	"listGameGenres",
	"listGameTrending",
	"listItems",
	"listMovieGenres",
	"listMovieTrending",
	"listSeriesGenres",
	"listSeriesTrending",
	"listsIndex",
	"listsShow",
	"me",
	"meShow",
	"notificationsIndex",
	"searchIndex",
	"searchSuggest",
	"usersShow",
]);

export const dehydrateOptions: DehydrateOptions = {
	shouldDehydrateMutation: () => false,
	// Success only, so an error state (and its RFC 9457 body) never reaches disk.
	shouldDehydrateQuery: (query) =>
		query.state.status === "success" && PERSISTED_QUERIES.has(String(query.queryKey[0])),
};

export function cacheBuster(appVersion: string, formulaVersion: number | null): string {
	return `${appVersion}+f${formulaVersion ?? 0}`;
}

const encoder = new TextEncoder();
const bytes = (value: string): number => encoder.encode(value).length;

// Drops the oldest queries until the serialized client fits the budget (NFR-10-6).
function trimToBudget(client: PersistedClient, maxBytes: number): string {
	const whole = JSON.stringify(client);
	if (bytes(whole) <= maxBytes) return whole;
	const queries = [...client.clientState.queries].sort(
		(a, b) => b.state.dataUpdatedAt - a.state.dataUpdatedAt,
	);
	const shell = bytes(
		JSON.stringify({ ...client, clientState: { ...client.clientState, queries: [] } }),
	);
	const kept: typeof queries = [];
	let total = shell;
	for (const query of queries) {
		const size = bytes(JSON.stringify(query)) + 1;
		if (total + size > maxBytes) break;
		kept.push(query);
		total += size;
	}
	return JSON.stringify({ ...client, clientState: { ...client.clientState, queries: kept } });
}

export function createSqlitePersister(storage: KeyValueStorage, maxBytes = MAX_BYTES): Persister {
	return {
		persistClient: (client) => storage.setItem(CACHE_KEY, trimToBudget(client, maxBytes)),
		async restoreClient() {
			const raw = await storage.getItem(CACHE_KEY);
			if (!raw) return undefined;
			try {
				return JSON.parse(raw) as PersistedClient;
			} catch {
				await storage.removeItem(CACHE_KEY);
				return undefined;
			}
		},
		removeClient: () => storage.removeItem(CACHE_KEY),
	};
}

export function readFormulaVersion(storage: KeyValueStorage): number | null {
	const value = Number(storage.getItemSync(FORMULA_KEY));
	return storage.getItemSync(FORMULA_KEY) !== null && Number.isInteger(value) ? value : null;
}

// The buster is fixed at launch; a new formulaVersion seen now busts the cache on the next launch.
export function trackFormulaVersion(client: QueryClient, storage: KeyValueStorage): () => void {
	let last = readFormulaVersion(storage);
	return client.getQueryCache().subscribe((event) => {
		if (event.type !== "updated" || event.action.type !== "success") return;
		// shortcut: reads the top-level formulaVersion (item payloads) only; extend if list rows start carrying it.
		const version: unknown = (event.query.state.data as { formulaVersion?: unknown } | undefined)
			?.formulaVersion;
		if (typeof version !== "number" || version === last) return;
		last = version;
		void storage.setItem(FORMULA_KEY, String(version));
	});
}
