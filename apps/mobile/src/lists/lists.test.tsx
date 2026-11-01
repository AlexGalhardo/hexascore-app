import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { act, fireEvent, render, screen, waitFor } from "@testing-library/react-native";
import type { ReactElement } from "react";
import { ApiError } from "../../../../packages/app-client/src/api";
import type { IdempotencyKey } from "../api/client";
import { setNetworkState, subscribeNetwork } from "../platform/network";
import type { KeyValueStorage } from "../query/persister";
import { startQueue } from "../queue/drain";
import type { SendOutcome } from "../queue/policy";
import type { QueueEntry } from "../queue/store";
import { createRatingRuntime } from "../rating/runtime";
import { AddToListSheet } from "./AddToListSheet";
import { type ListDetail, type ListsDeps, listKeys, listLaneResult } from "./data";
import { ListDetailScreen } from "./ListDetail";
import { ListsTab } from "./ListsTab";

function memoryStorage(): KeyValueStorage {
	const data = new Map<string, string>();
	return {
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

let counter = 0;
const newKey = (): IdempotencyKey => `key-${++counter}` as IdempotencyKey;

const summary = {
	id: "l1",
	type: "game" as const,
	title: "Backlog",
	isPublic: true,
	hidden: false,
};
const detail = (version: number, ids: string[]): ListDetail => ({
	...summary,
	description: null,
	itemCount: ids.length,
	version,
	items: ids.map((id, position) => ({
		position: position + 1,
		note: null,
		item: { id, type: "game", slug: id, title: `Title ${id}`, year: 2020, cover: null },
	})),
});

type Calls = { create: unknown[]; order: unknown[]; show: number };

function makeDeps(overrides: Partial<ListsDeps> = {}): ListsDeps & { calls: Calls } {
	const calls: Calls = { create: [], order: [], show: 0 };
	const client = {
		listsCreate: jest.fn(async ({ body }: { body?: unknown }) => {
			calls.create.push(body);
			return { id: "l2" };
		}),
		listsOrder: jest.fn(async ({ body }: { body?: unknown }) => {
			calls.order.push(body);
			return { version: 4 };
		}),
	};
	return {
		calls,
		read: {
			listsIndex: jest.fn(async () => ({ data: [{ ...summary, itemCount: 2, coverThumbs: [] }] })),
			listsShow: jest.fn(async () => {
				calls.show++;
				return detail(3, ["a", "b"]);
			}),
		},
		write: (_key, call) => call(client as never),
		newKey,
		queue: { enqueue: jest.fn(async () => null), drain: jest.fn(async () => {}) },
		notify: jest.fn(),
		...overrides,
	};
}

// Infinite gcTime: no garbage-collection timer keeps Jest alive after the run.
const testClient = (): QueryClient =>
	new QueryClient({
		defaultOptions: { queries: { retry: false, gcTime: Number.POSITIVE_INFINITY } },
	});

function wrap(node: ReactElement, client = testClient()): ReactElement {
	return <QueryClientProvider client={client}>{node}</QueryClientProvider>;
}

afterEach(() => setNetworkState("unknown"));

test("offline: create, rename, delete and reorder are disabled; remove stays enabled", async () => {
	setNetworkState("offline");
	const deps = makeDeps();
	await render(wrap(<ListsTab deps={deps} locale="en" onOpen={() => {}} />));
	expect(await screen.findByText("Backlog")).toBeTruthy();
	await fireEvent.changeText(screen.getByLabelText("Title"), "Weekend");
	expect(screen.getByRole("button", { name: "Create list" })).toBeDisabled();
	expect(screen.getByText(/Connect to the internet/)).toBeTruthy();

	await render(
		wrap(
			<ListDetailScreen
				deps={deps}
				id="l1"
				locale="en"
				onOpenItem={() => {}}
				onDeleted={() => {}}
			/>,
		),
	);
	expect(await screen.findByText("Title a")).toBeTruthy();
	expect(screen.getByRole("button", { name: "Edit list" })).toBeDisabled();
	expect(screen.getByRole("button", { name: "Delete list" })).toBeDisabled();
	expect(screen.getByRole("button", { name: "Move Title b up" })).toBeDisabled();
	expect(screen.getByRole("button", { name: "Remove Title a from the list" })).toBeEnabled();
});

test("online create sends the typed body", async () => {
	setNetworkState("online");
	const deps = makeDeps();
	await render(wrap(<ListsTab deps={deps} locale="en" onOpen={() => {}} />));
	await fireEvent.changeText(screen.getByLabelText("Title"), "  Weekend  ");
	await fireEvent.press(screen.getByRole("button", { name: "Create list" }));
	expect(deps.calls.create).toEqual([{ type: "game", title: "Weekend", isPublic: true }]);
});

test("offline add is queued on the list lane and drains on reconnect", async () => {
	setNetworkState("offline");
	let networkOnline = false;
	const sent: QueueEntry[] = [];
	const runtime = createRatingRuntime({
		storage: memoryStorage(),
		queryClient: testClient(),
		newKey,
		send: async (entry): Promise<SendOutcome> => {
			sent.push(entry);
			return { ok: true, body: {} };
		},
		online: () => networkOnline,
		paused: () => false,
		notify: () => {},
	});
	runtime.setUser("u1");
	const stop = startQueue(runtime.drain, {
		subscribeNetwork,
		subscribeAppActive: () => () => {},
	});
	const deps = makeDeps({ queue: runtime.queue });

	await render(wrap(<AddToListSheet deps={deps} itemId="it9" type="game" locale="en" />));
	await fireEvent.press(await screen.findByRole("button", { name: /Backlog/ }));
	expect(sent).toHaveLength(0);
	expect(deps.notify).toHaveBeenCalledWith("Added to Backlog.");

	networkOnline = true;
	await act(async () => setNetworkState("online"));
	await waitFor(() => expect(sent).toHaveLength(1));
	expect(sent[0]).toMatchObject({ op: "list:l1", itemId: "it9", payload: { method: "put" } });
	stop();
});

test("409 list-version-conflict on reorder refetches the list and shows a toast", async () => {
	setNetworkState("online");
	const deps = makeDeps({
		write: async () => {
			throw new ApiError(409, {
				status: 409,
				title: "Conflict",
				code: "list-version-conflict",
			} as never);
		},
	});
	await render(
		wrap(
			<ListDetailScreen
				deps={deps}
				id="l1"
				locale="en"
				onOpenItem={() => {}}
				onDeleted={() => {}}
			/>,
		),
	);
	await fireEvent.press(await screen.findByRole("button", { name: "Move Title b up" }));
	await waitFor(() => expect(deps.calls.show).toBe(2));
	expect(deps.notify).toHaveBeenCalledWith(
		"This list changed somewhere else. Showing the latest version.",
	);
});

test("reorder sends the full permutation with the list version", async () => {
	setNetworkState("online");
	const deps = makeDeps();
	await render(
		wrap(
			<ListDetailScreen
				deps={deps}
				id="l1"
				locale="en"
				onOpenItem={() => {}}
				onDeleted={() => {}}
			/>,
		),
	);
	await fireEvent.press(await screen.findByRole("button", { name: "Move Title b up" }));
	expect(deps.calls.order).toEqual([{ version: 3, itemIds: ["b", "a"] }]);
});

test("a rolled-back list write refetches that list and toasts", () => {
	const client = testClient();
	const invalidate = jest.spyOn(client, "invalidateQueries");
	const notify = jest.fn();
	const entry = { op: "list:l1", itemId: "a" } as unknown as QueueEntry;
	listLaneResult(client, notify, "en")(entry, { kind: "rollback", code: null, retryAfter: null });
	expect(invalidate).toHaveBeenCalledWith({ queryKey: listKeys.one("l1") });
	expect(notify).toHaveBeenCalledWith("Your list change was not saved.");
	listLaneResult(client, notify, "en")({ ...entry, op: "rating" }, { kind: "applied", body: {} });
	expect(notify).toHaveBeenCalledTimes(1);
});
