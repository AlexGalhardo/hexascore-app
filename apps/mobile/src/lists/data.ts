import { type QueryClient, useQuery, useQueryClient } from "@tanstack/react-query";
import { useSyncExternalStore } from "react";
import { ApiError } from "../../../../packages/app-client/src/api";
import type { AppClient, IdempotencyKey, ReadClient } from "../api/client";
import { type AppLocale, t } from "../i18n";
import { getNetworkState, subscribeNetwork } from "../platform/network";
import type { QueueResult } from "../queue/drain";
import type { QueueEntry } from "../queue/store";
import type { RatingQueue } from "../rating/useRateMutation";

/** Server list type (`requests.ListCreateSchema`). */
export type ListType = "game" | "movie" | "series" | "book";

export type ListSummary = {
	id: string;
	type: ListType;
	title: string;
	isPublic: boolean;
	hidden: boolean;
	itemCount: number;
	coverThumbs: string[];
};

export type ListEntry = {
	position: number;
	note: string | null;
	item: {
		id: string;
		type: string;
		slug: string;
		title: string;
		year: number | null;
		cover: { thumb: string; full: string } | null;
	};
};

export type ListDetail = Omit<ListSummary, "itemCount" | "coverThumbs"> & {
	description: string | null;
	itemCount: number;
	version: number;
	items: ListEntry[];
};

export type ListsDeps = {
	read: Pick<ReadClient, "listsIndex" | "listsShow">;
	write: AppClient["write"];
	newKey: () => IdempotencyKey;
	/** The shared offline queue: a single add/remove rides the `list:<id>` lane (SDD 10 §7.3). */
	queue: Pick<RatingQueue, "enqueue" | "drain">;
	notify: (message: string) => void;
};

// Element 0 is the ReadClient method, as in query/keys.ts.
export const listKeys = {
	mine: ["listsIndex"] as const,
	one: (id: string) => ["listsShow", { id }] as const,
};

const LANE = "list:";

export function useOnline(): boolean {
	return useSyncExternalStore(subscribeNetwork, getNetworkState) !== "offline";
}

export function useMyLists(deps: ListsDeps) {
	return useQuery({
		queryKey: listKeys.mine,
		queryFn: async () => ((await deps.read.listsIndex()) as { data?: ListSummary[] }).data ?? [],
	});
}

export function useList(deps: ListsDeps, id: string) {
	return useQuery({
		queryKey: listKeys.one(id),
		queryFn: async () => (await deps.read.listsShow({ params: { id } })) as ListDetail,
	});
}

const problemCode = (error: unknown): string | null => {
	if (!(error instanceof ApiError)) return null;
	const code: unknown = (error.problem as { code?: unknown } | undefined)?.code;
	return typeof code === "string" ? code : null;
};

// Reorder conflicts: the server wins, so the screen refetches (SDD 14 §7.1).
const STALE = new Set(["list-version-conflict", "order-mismatch"]);

export type ListActions = {
	create: (input: { type: ListType; title: string; firstItemId?: string }) => Promise<boolean>;
	rename: (id: string, title: string) => Promise<boolean>;
	remove: (id: string) => Promise<boolean>;
	move: (list: ListDetail, itemId: string, delta: -1 | 1) => Promise<boolean>;
	/** Queued: works offline. */
	addItem: (list: Pick<ListSummary, "id" | "title">, itemId: string) => Promise<boolean>;
	/** Queued: works offline; the row leaves the cached list at once. */
	removeItem: (listId: string, itemId: string) => Promise<boolean>;
};

export function useListActions(deps: ListsDeps, locale: AppLocale): ListActions {
	const client = useQueryClient();
	const refresh = (id?: string): void => {
		void client.invalidateQueries({ queryKey: listKeys.mine });
		if (id) void client.invalidateQueries({ queryKey: listKeys.one(id) });
	};

	// Create, rename, delete and reorder are online only; the screens disable them offline.
	const online = async (id: string | undefined, call: Parameters<AppClient["write"]>[1]) => {
		try {
			await deps.write(deps.newKey(), call);
			refresh(id);
			return true;
		} catch (error) {
			const code = problemCode(error);
			deps.notify(
				t(locale, code !== null && STALE.has(code) ? "lists.conflict" : "lists.notSaved"),
			);
			refresh(id);
			return false;
		}
	};

	const queued = async (
		listId: string,
		itemId: string,
		method: "put" | "delete",
	): Promise<boolean> => {
		const entry = await deps.queue.enqueue({
			op: `${LANE}${listId}`,
			itemId,
			payload: method === "put" ? { method, body: {} } : { method },
		});
		if (entry === null) {
			deps.notify(t(locale, "lists.notSaved"));
			return false;
		}
		void deps.queue.drain();
		return true;
	};

	return {
		create: ({ type, title, firstItemId }) =>
			online(undefined, (c) =>
				c.listsCreate({
					body: {
						type,
						title: title.trim(),
						isPublic: true,
						...(firstItemId ? { firstItemId } : {}),
					},
				}),
			),
		rename: (id, title) =>
			online(id, (c) => c.listsUpdate({ params: { id }, body: { title: title.trim() } })),
		async remove(id) {
			const ok = await online(undefined, (c) => c.listsDelete({ params: { id } }));
			if (ok) client.removeQueries({ queryKey: listKeys.one(id) });
			return ok;
		},
		move(list, itemId, delta) {
			const ids = list.items.map((e) => e.item.id);
			const from = ids.indexOf(itemId);
			const to = from + delta;
			if (from < 0 || to < 0 || to >= ids.length) return Promise.resolve(false);
			[ids[from], ids[to]] = [ids[to] as string, ids[from] as string];
			return online(list.id, (c) =>
				c.listsOrder({ params: { id: list.id }, body: { version: list.version, itemIds: ids } }),
			);
		},
		async addItem(list, itemId) {
			const ok = await queued(list.id, itemId, "put");
			if (ok) deps.notify(t(locale, "lists.addedTo", { title: list.title }));
			return ok;
		},
		async removeItem(listId, itemId) {
			const before = client.getQueryData<ListDetail>(listKeys.one(listId));
			if (before) {
				const items = before.items.filter((e) => e.item.id !== itemId);
				client.setQueryData(listKeys.one(listId), { ...before, items, itemCount: items.length });
			}
			const ok = await queued(listId, itemId, "delete");
			if (!ok && before) client.setQueryData(listKeys.one(listId), before);
			return ok;
		},
	};
}

/** The drain's onResult for the `list:<id>` lane: the server is the truth after every write. */
export function listLaneResult(
	client: QueryClient,
	notify: (message: string) => void,
	locale: () => AppLocale,
): (entry: QueueEntry, result: QueueResult) => void {
	return (entry, result) => {
		if (!entry.op.startsWith(LANE)) return;
		void client.invalidateQueries({ queryKey: listKeys.one(entry.op.slice(LANE.length)) });
		void client.invalidateQueries({ queryKey: listKeys.mine });
		if (result.kind === "rollback" || (result.kind === "drop" && result.toast)) {
			notify(t(locale(), "lists.notSaved"));
		}
	};
}
