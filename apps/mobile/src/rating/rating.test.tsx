import { QueryClient } from "@tanstack/react-query";
import { act, fireEvent, render, renderHook, screen } from "@testing-library/react-native";
import * as Haptics from "expo-haptics";
import type { ReactNode } from "react";
import { queryKeys } from "../query/keys";
import type { QueuedWrite, QueueEntry } from "../queue/store";
import { createStatusCache } from "../status/cache";
import { createRatingCache } from "./cache";
import { StarInput } from "./StarInput";
import { type RatingQueue, RatingQueueProvider, useRateMutation } from "./useRateMutation";

jest.mock("expo-haptics", () => ({ selectionAsync: jest.fn(() => Promise.resolve()) }));

const ITEM = "0190c0de-0000-7000-8000-000000000001";
const itemKey = queryKeys.item("games", ITEM, "en");
const meKey = queryKeys.itemMe(ITEM);

function seeded(): QueryClient {
	const client = new QueryClient({
		defaultOptions: { queries: { retry: false, gcTime: Number.POSITIVE_INFINITY } },
	});
	client.setQueryData(itemKey, { id: ITEM, score: 70, ratingCount: 4, confidence: null });
	client.setQueryData(meKey, { rating: null, status: "playing", timeToBeat: null });
	return client;
}

function entryOf(write: QueuedWrite): QueueEntry {
	return { ...write, id: "k1", idempotencyKey: "k1" as never, createdAt: 0, attempts: 0 };
}

async function setup(full = false) {
	const client = seeded();
	const cache = createRatingCache(client);
	const writes: QueuedWrite[] = [];
	const queue: RatingQueue = {
		cache,
		statusCache: createStatusCache(client),
		enqueue: jest.fn(async (write: QueuedWrite) => {
			if (full) return null;
			writes.push(write);
			return entryOf(write);
		}),
		drain: jest.fn(() => Promise.resolve()),
	};
	const wrapper = ({ children }: { children: ReactNode }) => (
		<RatingQueueProvider value={queue}>{children}</RatingQueueProvider>
	);
	const { result } = await renderHook(() => useRateMutation(ITEM), { wrapper });
	return { client, cache, queue, writes, mutation: result };
}

describe("StarInput", () => {
	beforeEach(() => jest.mocked(Haptics.selectionAsync).mockClear());

	async function mount(value: number | null) {
		const onChange = jest.fn();
		await render(<StarInput label="Gameplay" value={value} onChange={onChange} locale="en" />);
		const stars = screen.getByTestId("rating-stars", { includeHiddenElements: true });
		await fireEvent(stars, "layout", { nativeEvent: { layout: { width: 300 } } });
		return { onChange, stars };
	}

	test("a tap on the left half of star 7 gives 6.5, with one selection haptic", async () => {
		const { onChange, stars } = await mount(null);
		await fireEvent.press(stars, { nativeEvent: { locationX: 185 } });
		expect(onChange).toHaveBeenCalledWith(6.5);
		expect(Haptics.selectionAsync).toHaveBeenCalledTimes(1);
	});

	test("a re-tap of the current value is a no-op", async () => {
		const { onChange, stars } = await mount(6.5);
		await fireEvent.press(stars, { nativeEvent: { locationX: 185 } });
		expect(onChange).not.toHaveBeenCalled();
		expect(Haptics.selectionAsync).not.toHaveBeenCalled();
	});

	test("the accessibility increment from 6.5 gives 7.0", async () => {
		const { onChange } = await mount(6.5);
		const slider = screen.getByRole("adjustable", { name: "Gameplay" });
		await fireEvent(slider, "accessibilityAction", { nativeEvent: { actionName: "increment" } });
		expect(onChange).toHaveBeenCalledWith(7);
		expect(Haptics.selectionAsync).toHaveBeenCalledTimes(1);
	});

	test("an increment at the top end changes nothing", async () => {
		await render(<StarInput label="Story" value={10} onChange={jest.fn()} locale="en" />);
		const slider = screen.getByRole("adjustable", { name: "Story" });
		await fireEvent(slider, "accessibilityAction", { nativeEvent: { actionName: "increment" } });
		expect(Haptics.selectionAsync).not.toHaveBeenCalled();
	});
});

describe("useRateMutation", () => {
	test("rate enqueues one PUT with every key and updates the cache at once", async () => {
		const { client, queue, writes, mutation } = await setup();
		await act(() => mutation.current.rate({ gameplay: 8, story: 6.5 }));
		expect(writes).toEqual([
			{
				op: "rating",
				itemId: ITEM,
				payload: { method: "put", body: { scores: { gameplay: 8, story: 6.5 } } },
			},
		]);
		expect(queue.drain).toHaveBeenCalledTimes(1);
		expect(client.getQueryData(meKey)).toMatchObject({
			rating: { gameplay: 8, story: 6.5 },
			status: "playing",
		});
		expect(client.getQueryData(itemKey)).toMatchObject({ ratingCount: 5 });
	});

	test("the server body replaces the optimistic aggregate on 2xx", async () => {
		const { client, cache, mutation } = await setup();
		await act(() => mutation.current.rate({ gameplay: 8 }));
		const write: QueuedWrite = {
			op: "rating",
			itemId: ITEM,
			payload: { method: "put", body: { scores: { gameplay: 8 } } },
		};
		cache.onResult(entryOf(write), {
			kind: "applied",
			body: {
				rating: { gameplay: 8 },
				item: { score: 72, scoreByCategory: { gameplay: 80 }, ratingCount: 5, confidence: "low" },
			},
		});
		expect(client.getQueryData(itemKey)).toMatchObject({
			score: 72,
			ratingCount: 5,
			confidence: "low",
		});
		expect(client.getQueryData(meKey)).toMatchObject({ rating: { gameplay: 8 } });
	});

	test("409 rating-not-released rolls back to the state before the first write", async () => {
		const { client, cache, mutation } = await setup();
		await act(() => mutation.current.rate({ gameplay: 8 }));
		await act(() => mutation.current.rate({ gameplay: 9 }));
		const write: QueuedWrite = {
			op: "rating",
			itemId: ITEM,
			payload: { method: "put", body: { scores: { gameplay: 9 } } },
		};
		cache.onResult(entryOf(write), {
			kind: "rollback",
			code: "rating-not-released",
			retryAfter: null,
		});
		expect(client.getQueryData(meKey)).toMatchObject({ rating: null, status: "playing" });
		expect(client.getQueryData(itemKey)).toMatchObject({ score: 70, ratingCount: 4 });
	});

	test("remove enqueues a DELETE and drops the count; a 204 clears the rating", async () => {
		const { client, cache, writes, mutation } = await setup();
		client.setQueryData(meKey, { rating: { gameplay: 7 }, status: null, timeToBeat: null });
		await act(() => mutation.current.remove());
		expect(writes).toEqual([{ op: "rating", itemId: ITEM, payload: { method: "delete" } }]);
		expect(client.getQueryData(itemKey)).toMatchObject({ ratingCount: 3 });
		cache.onResult(entryOf(writes[0] as QueuedWrite), { kind: "applied", body: undefined });
		expect(client.getQueryData(meKey)).toMatchObject({ rating: null });
	});

	test("a full queue rolls back at once and reports false", async () => {
		const { client, mutation } = await setup(true);
		let accepted = true;
		await act(async () => {
			accepted = await mutation.current.rate({ gameplay: 8 });
		});
		expect(accepted).toBe(false);
		expect(client.getQueryData(meKey)).toMatchObject({ rating: null });
		expect(client.getQueryData(itemKey)).toMatchObject({ ratingCount: 4 });
	});

	test("results for other lanes are ignored", async () => {
		const { client, cache } = await setup();
		cache.onResult(entryOf({ op: "status", itemId: ITEM, payload: { method: "delete" } }), {
			kind: "rollback",
			code: null,
			retryAfter: null,
		});
		expect(client.getQueryData(meKey)).toMatchObject({ status: "playing" });
	});
});
