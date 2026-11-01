import type { Query, QueryClient, QueryKey } from "@tanstack/react-query";
import { queryKeys } from "../query/keys";
import type { QueueResult } from "../queue/drain";
import type { QueueEntry } from "../queue/store";

/** Stars per A-06 category key (0.5–10); the server multiplies by 10. */
export type Scores = Record<string, number>;

type Snapshot = { me: unknown; items: [QueryKey, unknown][] };

export type RatingCache = {
	/** Writes the user's rating (null = removed) into the cache before the queue sends it. */
	optimistic(itemId: string, scores: Scores | null): void;
	/** Puts back the state from before the first pending write of the item. */
	restore(itemId: string): void;
	/** The drain's onResult for the `rating` lane; other lanes are ignored. */
	onResult(entry: QueueEntry, result: QueueResult): void;
};

const SUMMARY_FIELDS = ["score", "scoreByCategory", "ratingCount", "confidence", "formulaVersion"];

const asObject = (value: unknown): Record<string, unknown> =>
	typeof value === "object" && value !== null ? (value as Record<string, unknown>) : {};

const hasRating = (me: unknown): boolean => Object.keys(asObject(asObject(me).rating)).length > 0;

const itemFilter = (itemId: string) => ({
	predicate: (query: Query) =>
		query.queryKey[0] === "getItem" && asObject(query.queryKey[1]).id === itemId,
});

export function createRatingCache(client: QueryClient): RatingCache {
	// Keyed by item: a newer write replaces the queued one, so only the first snapshot is the real "before".
	const snapshots = new Map<string, Snapshot>();

	const setItems = (itemId: string, patch: (item: Record<string, unknown>) => object): void => {
		client.setQueriesData(itemFilter(itemId), (item: unknown) =>
			typeof item === "object" && item !== null ? patch(asObject(item)) : item,
		);
	};

	const setRating = (itemId: string, rating: unknown): void => {
		client.setQueryData(queryKeys.itemMe(itemId), (me: unknown) => ({ ...asObject(me), rating }));
	};

	const restore = (itemId: string): void => {
		const snapshot = snapshots.get(itemId);
		snapshots.delete(itemId);
		if (snapshot) {
			client.setQueryData(queryKeys.itemMe(itemId), snapshot.me);
			for (const [key, data] of snapshot.items) client.setQueryData(key, data);
		}
		// Without a snapshot (the write outlived an app restart) the server is the only truth.
		void client.invalidateQueries({ queryKey: queryKeys.itemMe(itemId) });
		void client.invalidateQueries(itemFilter(itemId));
	};

	return {
		optimistic(itemId, scores) {
			const me = client.getQueryData(queryKeys.itemMe(itemId));
			if (!snapshots.has(itemId)) {
				snapshots.set(itemId, { me, items: client.getQueriesData(itemFilter(itemId)) });
			}
			const delta = (scores === null ? 0 : 1) - (hasRating(me) ? 1 : 0);
			setRating(itemId, scores);
			if (delta !== 0) {
				setItems(itemId, (item) =>
					typeof item.ratingCount === "number"
						? { ...item, ratingCount: item.ratingCount + delta }
						: item,
				);
			}
		},
		restore,
		onResult(entry, result) {
			if (entry.op !== "rating") return;
			if (result.kind !== "applied") {
				restore(entry.itemId);
				return;
			}
			snapshots.delete(entry.itemId);
			if (entry.payload.method === "delete") {
				setRating(entry.itemId, null);
				return;
			}
			const body = asObject(result.body);
			setRating(entry.itemId, body.rating);
			const summary = asObject(body.item);
			const fields = Object.fromEntries(
				SUMMARY_FIELDS.filter((f) => f in summary).map((f) => [f, summary[f]]),
			);
			setItems(entry.itemId, (item) => ({ ...item, ...fields }));
		},
	};
}
