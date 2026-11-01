import { QueryClient } from "@tanstack/react-query";
import { act, fireEvent, render, screen } from "@testing-library/react-native";
import type { ReactNode } from "react";
import type { IdempotencyKey } from "../api/client";
import type { SessionState } from "../auth/session";
import { queryKeys } from "../query/keys";
import type { KeyValueStorage } from "../query/persister";
import type { SendOutcome } from "../queue/policy";
import type { QueueEntry } from "../queue/store";
import { CATEGORY_KEYS } from "./categories";
import { RateSheet } from "./RateSheet";
import { createRatingRuntime, followMe } from "./runtime";
import { RatingQueueProvider } from "./useRateMutation";

jest.mock("expo-haptics", () => ({ selectionAsync: jest.fn(() => Promise.resolve()) }));

const ITEM = "0190c0de-0000-7000-8000-000000000002";
const itemKey = queryKeys.item("games", ITEM, "en");
const meKey = queryKeys.itemMe(ITEM);

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

function setup(outcome: (entry: QueueEntry) => SendOutcome) {
	const client = new QueryClient({
		defaultOptions: { queries: { retry: false, gcTime: Number.POSITIVE_INFINITY } },
	});
	client.setQueryData(itemKey, { id: ITEM, score: 70, ratingCount: 4, confidence: null });
	client.setQueryData(meKey, { rating: null, status: null, timeToBeat: null });
	let key = 0;
	let release: () => void = () => {};
	const gate = new Promise<void>((resolve) => {
		release = resolve;
	});
	const sent: QueueEntry[] = [];
	const notify = jest.fn();
	const runtime = createRatingRuntime({
		storage: memoryStorage(),
		queryClient: client,
		newKey: () => `key-${++key}` as IdempotencyKey,
		send: async (entry) => {
			sent.push(entry);
			await gate;
			return outcome(entry);
		},
		online: () => true,
		paused: () => false,
		notify,
	});
	runtime.setUser("42");
	const wrapper = ({ children }: { children: ReactNode }) => (
		<RatingQueueProvider value={runtime.queue}>{children}</RatingQueueProvider>
	);
	return { client, runtime, sent, notify, release, wrapper };
}

async function rateAll(value: number): Promise<void> {
	for (const slider of screen.getAllByRole("adjustable")) {
		for (let i = 0; i < value * 2; i++) {
			await fireEvent(slider, "accessibilityAction", { nativeEvent: { actionName: "increment" } });
		}
	}
}

const isDisabled = (label: string): boolean =>
	Boolean(screen.getByRole("button", { name: label }).props.accessibilityState?.disabled);

describe("RateSheet", () => {
	test("six labelled widgets; submit stays disabled until every category is set", async () => {
		const { wrapper } = setup(() => ({ ok: true, body: {} }));
		await render(<RateSheet itemId={ITEM} type="games" initial={null} locale="en" />, { wrapper });
		expect(screen.getAllByRole("adjustable")).toHaveLength(6);
		expect(screen.getByRole("adjustable", { name: "Gameplay" })).toBeTruthy();
		expect(isDisabled("Save rating")).toBe(true);
		const [first] = screen.getAllByRole("adjustable");
		await fireEvent(first as never, "accessibilityAction", {
			nativeEvent: { actionName: "increment" },
		});
		expect(isDisabled("Save rating")).toBe(true);
		await rateAll(4);
		expect(isDisabled("Save rating")).toBe(false);
		expect(screen.queryByRole("button", { name: "Delete rating" })).toBeNull();
	});

	test("the optimistic aggregate updates at once and the server body replaces it on 2xx", async () => {
		const { client, sent, release, wrapper } = setup((entry) => ({
			ok: true,
			body: {
				rating: (entry.payload as { body: { scores: Record<string, number> } }).body.scores,
				item: { score: 76, ratingCount: 5, confidence: "low" },
			},
		}));
		await render(<RateSheet itemId={ITEM} type="games" initial={null} locale="en" />, { wrapper });
		await rateAll(4);
		await fireEvent.press(screen.getByRole("button", { name: "Save rating" }));
		expect(client.getQueryData(itemKey)).toMatchObject({ score: 70, ratingCount: 5 });
		expect(sent).toHaveLength(1);
		expect(sent[0]?.payload).toEqual({
			method: "put",
			body: { scores: Object.fromEntries(CATEGORY_KEYS.games.map((k) => [k, 4])) },
		});
		await act(async () => {
			release();
			await Promise.resolve();
		});
		await act(() => Promise.resolve());
		expect(client.getQueryData(itemKey)).toMatchObject({
			score: 76,
			ratingCount: 5,
			confidence: "low",
		});
	});

	test("409 rating-not-released rolls back and toasts", async () => {
		const { client, notify, release, runtime, wrapper } = setup(() => ({
			ok: false,
			status: 409,
			code: "rating-not-released",
			retryAfter: null,
		}));
		await render(<RateSheet itemId={ITEM} type="games" initial={null} locale="en" />, { wrapper });
		await rateAll(3);
		await fireEvent.press(screen.getByRole("button", { name: "Save rating" }));
		expect(client.getQueryData(meKey)).toMatchObject({ rating: { gameplay: 3 } });
		await act(async () => {
			release();
			await runtime.drain();
		});
		expect(client.getQueryData(meKey)).toMatchObject({ rating: null });
		expect(client.getQueryData(itemKey)).toMatchObject({ ratingCount: 4 });
		expect(notify).toHaveBeenCalledWith("ratings.notReleased");
	});

	test("remove-rating sends one DELETE through the queue", async () => {
		const { sent, release, runtime, wrapper } = setup(() => ({ ok: true, body: undefined }));
		const initial = Object.fromEntries(CATEGORY_KEYS.games.map((k) => [k, 7]));
		await render(<RateSheet itemId={ITEM} type="games" initial={initial} locale="en" />, {
			wrapper,
		});
		expect(isDisabled("Update rating")).toBe(false);
		await fireEvent.press(screen.getByRole("button", { name: "Delete rating" }));
		await act(async () => {
			release();
			await runtime.drain();
		});
		expect(sent.map((e) => e.payload)).toEqual([{ method: "delete" }]);
	});

	test("without a user nothing is queued and the sheet says so", async () => {
		const { runtime, sent, wrapper } = setup(() => ({ ok: true, body: {} }));
		runtime.setUser(null);
		await render(<RateSheet itemId={ITEM} type="games" initial={null} locale="en" />, { wrapper });
		await rateAll(2);
		await fireEvent.press(screen.getByRole("button", { name: "Save rating" }));
		expect(sent).toHaveLength(0);
		expect(screen.getByText("Your rating could not be saved. Try again.")).toBeTruthy();
	});
});

describe("followMe", () => {
	function fakeSession() {
		let listener: (state: SessionState) => void = () => {};
		return {
			observeMe: jest.fn(),
			handleError: jest.fn(async () => false),
			subscribe: (next: (state: SessionState) => void) => {
				listener = next;
				return () => {};
			},
			emit: (state: SessionState) => listener(state),
		};
	}

	test("an active session reads /me, feeds observeMe and binds the queue to the id", async () => {
		const session = fakeSession();
		const setUser = jest.fn();
		const me = { id: "42", deletionScheduledAt: null };
		followMe(session, async () => me, setUser);
		session.emit({ kind: "active", emailNotVerified: false });
		await act(() => Promise.resolve());
		expect(session.observeMe).toHaveBeenCalledWith(me);
		expect(setUser).toHaveBeenLastCalledWith("42");
		session.emit({ kind: "signedOut" });
		expect(setUser).toHaveBeenLastCalledWith(null);
	});

	test("a failed /me goes to the session error handler", async () => {
		const session = fakeSession();
		const error = new Error("offline");
		followMe(
			session,
			async () => {
				throw error;
			},
			jest.fn(),
		);
		session.emit({ kind: "active", emailNotVerified: false });
		await act(() => Promise.resolve());
		expect(session.handleError).toHaveBeenCalledWith(error);
	});
});
