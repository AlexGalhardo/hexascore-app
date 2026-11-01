import type { QueryClient } from "@tanstack/react-query";
import { queryKeys } from "../query/keys";
import type { QueueResult } from "../queue/drain";
import type { QueueEntry } from "../queue/store";

export const STATUSES = ["want", "in_progress", "completed", "dropped"] as const;
export type ItemStatus = (typeof STATUSES)[number];

export const STYLES = ["mainHours", "extrasHours", "completionistHours"] as const;
export type TimeToBeat = Partial<Record<(typeof STYLES)[number], number>>;

type Field = "status" | "timeToBeat";

export type StatusCache = {
	/** Writes the user's status or time-to-beat (null = removed) into `/items/{id}/me` before the queue sends it. */
	optimistic(itemId: string, field: Field, value: unknown): void;
	/** Puts back the value from before the first pending write of that field. */
	restore(itemId: string, field: Field): void;
	/** The drain's onResult for the `status` and `timeToBeat` lanes; other lanes are ignored. */
	onResult(entry: QueueEntry, result: QueueResult): void;
};

const asObject = (value: unknown): Record<string, unknown> =>
	typeof value === "object" && value !== null ? (value as Record<string, unknown>) : {};

export function statusOf(me: unknown): ItemStatus | null {
	const status = asObject(me).status;
	return STATUSES.find((s) => s === status) ?? null;
}

export function timeToBeatOf(me: unknown): TimeToBeat | null {
	const raw = asObject(me).timeToBeat;
	if (typeof raw !== "object" || raw === null) return null;
	const out: TimeToBeat = {};
	for (const style of STYLES) {
		const hours = asObject(raw)[style];
		if (typeof hours === "number") out[style] = hours;
	}
	return out;
}

/** Hours as typed (dot or comma); null when empty, "invalid" outside 0.5–10,000 or off the 0.5 step (SDD 02 §6). */
export function parseHours(text: string): number | null | "invalid" {
	const trimmed = text.trim().replace(",", ".");
	if (trimmed === "") return null;
	const hours = Number(trimmed);
	if (!Number.isFinite(hours) || hours < 0.5 || hours > 10_000 || (hours * 2) % 1 !== 0) {
		return "invalid";
	}
	return hours;
}

const fieldOf = (entry: QueueEntry): Field | null =>
	entry.op === "status" || entry.op === "timeToBeat" ? entry.op : null;

export function createStatusCache(client: QueryClient): StatusCache {
	// Keyed by item and field: a newer write replaces the queued one, so only the first snapshot is the real "before".
	const snapshots = new Map<string, unknown>();
	const slot = (itemId: string, field: Field): string => `${field}:${itemId}`;

	const set = (itemId: string, field: Field, value: unknown): void => {
		client.setQueryData(queryKeys.itemMe(itemId), (me: unknown) => ({
			...asObject(me),
			[field]: value,
		}));
	};

	const restore = (itemId: string, field: Field): void => {
		const key = slot(itemId, field);
		if (snapshots.has(key)) set(itemId, field, snapshots.get(key));
		snapshots.delete(key);
		void client.invalidateQueries({ queryKey: queryKeys.itemMe(itemId) });
	};

	return {
		optimistic(itemId, field, value) {
			const key = slot(itemId, field);
			if (!snapshots.has(key)) {
				snapshots.set(key, asObject(client.getQueryData(queryKeys.itemMe(itemId)))[field] ?? null);
			}
			set(itemId, field, value);
		},
		restore,
		onResult(entry, result) {
			const field = fieldOf(entry);
			if (field === null) return;
			if (result.kind !== "applied") {
				restore(entry.itemId, field);
				return;
			}
			snapshots.delete(slot(entry.itemId, field));
			const value = entry.payload.method === "delete" ? null : asObject(result.body)[field];
			if (value !== undefined) set(entry.itemId, field, value);
		},
	};
}
