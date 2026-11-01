import { type QueryClient, useQuery, useQueryClient } from "@tanstack/react-query";
import type { Client } from "../../../../packages/app-client/src/api";
import type { ReadClient } from "../api/client";

export type Notification = {
	id: string;
	kind: string;
	readAt: string | null;
	createdAt: string;
};

export type NotificationsPage = { data: Notification[]; unread: number };

export type NotificationsApi = {
	read: Pick<ReadClient, "notificationsIndex">;
	write<T>(call: (client: Pick<Client, "notificationsRead">) => Promise<T>): Promise<T>;
};

export const notificationsKey = ["notificationsIndex"] as const;

/** SDD 14 §6.3 caps POST /me/notifications/read at 100 ids. */
export const MAX_READ_IDS = 100;

const str = (value: unknown): string => (typeof value === "string" ? value : "");

// The generated client types the list as unknown; narrow at the boundary.
export function parsePage(body: unknown): NotificationsPage {
	const page = (typeof body === "object" && body !== null ? body : {}) as {
		data?: unknown;
		meta?: { unread?: unknown };
	};
	const rows = Array.isArray(page.data) ? page.data : [];
	return {
		data: rows
			.filter((row): row is Record<string, unknown> => typeof row === "object" && row !== null)
			.map((row) => ({
				id: str(row.id),
				kind: str(row.kind),
				readAt: typeof row.readAt === "string" ? row.readAt : null,
				createdAt: str(row.createdAt),
			})),
		unread: typeof page.meta?.unread === "number" ? page.meta.unread : 0,
	};
}

export function useNotifications(api: NotificationsApi, enabled = true) {
	return useQuery({
		queryKey: notificationsKey,
		queryFn: async () => parsePage(await api.read.notificationsIndex()),
		enabled,
	});
}

/** The badge reads `meta.unread` from the list response; nothing polls (ROADMAP P5.B.4.3). */
export function useUnread(api: NotificationsApi, enabled = true): number {
	return useNotifications(api, enabled).data?.unread ?? 0;
}

/** Marks ids read in requests of at most 100, then updates the cached page and badge without a refetch. */
export async function markRead(
	api: NotificationsApi,
	client: QueryClient,
	ids: string[],
): Promise<void> {
	for (let i = 0; i < ids.length; i += MAX_READ_IDS) {
		const batch = ids.slice(i, i + MAX_READ_IDS);
		await api.write((c) => c.notificationsRead({ body: { ids: batch } }));
		const read = new Set(batch);
		const at = new Date().toISOString();
		client.setQueryData<NotificationsPage>(notificationsKey, (page) => {
			if (page === undefined) return page;
			const fresh = page.data.filter((n) => read.has(n.id) && n.readAt === null).length;
			return {
				data: page.data.map((n) =>
					read.has(n.id) && n.readAt === null ? { ...n, readAt: at } : n,
				),
				unread: Math.max(0, page.unread - fresh),
			};
		});
	}
}

export function useMarkRead(api: NotificationsApi): (ids: string[]) => Promise<void> {
	const client = useQueryClient();
	return (ids) => markRead(api, client, ids);
}
