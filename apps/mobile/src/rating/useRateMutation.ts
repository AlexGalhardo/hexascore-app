import { createContext, useCallback, useContext, useMemo } from "react";
import type { QueuedWrite, QueueEntry } from "../queue/store";
import type { RatingCache, Scores } from "./cache";

/** The offline queue as the rating UI sees it; runtime binds `enqueue` to the `/me` user (P5.B.2.3). */
export type RatingQueue = {
	cache: RatingCache;
	enqueue: (write: QueuedWrite) => Promise<QueueEntry | null>;
	drain: () => Promise<void>;
};

const RatingQueueContext = createContext<RatingQueue | null>(null);

export const RatingQueueProvider = RatingQueueContext.Provider;

export type RateMutation = {
	/** One PUT with every category key. False when the queue is full and the change was rolled back. */
	rate: (scores: Scores) => Promise<boolean>;
	remove: () => Promise<boolean>;
};

export function useRateMutation(itemId: string): RateMutation {
	const queue = useContext(RatingQueueContext);
	if (queue === null) throw new Error("useRateMutation needs a RatingQueueProvider");

	const send = useCallback(
		async (scores: Scores | null): Promise<boolean> => {
			queue.cache.optimistic(itemId, scores);
			const entry = await queue.enqueue({
				op: "rating",
				itemId,
				payload: scores === null ? { method: "delete" } : { method: "put", body: { scores } },
			});
			if (entry === null) {
				queue.cache.restore(itemId);
				return false;
			}
			void queue.drain();
			return true;
		},
		[queue, itemId],
	);

	return useMemo(() => ({ rate: (scores) => send(scores), remove: () => send(null) }), [send]);
}
