import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { fireEvent, render, screen, waitFor } from "@testing-library/react-native";
import type { ReactElement } from "react";
import { queryKeys } from "../query/keys";
import type { KeyValueStorage } from "../query/persister";
import type { SendOutcome } from "../queue/policy";
import type { QueueEntry } from "../queue/store";
import { createRatingRuntime } from "../rating/runtime";
import { RatingQueueProvider } from "../rating/useRateMutation";
import { type ItemReader, ItemView } from "../screens/item/Item";
import { parseHours, statusOf, timeToBeatOf } from "./cache";
import { StatusSheet } from "./StatusSheet";
import { TimeToBeatSheet } from "./TimeToBeatSheet";

function memoryStorage(): KeyValueStorage {
	const data = new Map<string, string>();
	return {
		getItem: async (key: string) => data.get(key) ?? null,
		getItemSync: (key: string) => data.get(key) ?? null,
		setItem: async (key: string, value: string) => void data.set(key, value),
		removeItem: async (key: string) => void data.delete(key),
	};
}

function setup(send: (entry: QueueEntry) => Promise<SendOutcome>) {
	const queryClient = new QueryClient();
	const notify = jest.fn();
	let n = 0;
	const runtime = createRatingRuntime({
		storage: memoryStorage(),
		queryClient,
		newKey: () => `key-${++n}` as never,
		send,
		online: () => true,
		paused: () => false,
		notify,
		schedule: () => () => {},
	});
	runtime.setUser("7");
	const wrap = (ui: ReactElement) => (
		<RatingQueueProvider value={runtime.queue}>{ui}</RatingQueueProvider>
	);
	return { queryClient, notify, wrap };
}

const ok = (body: unknown) => async (): Promise<SendOutcome> => ({ ok: true, body });

describe("parseHours", () => {
	test("accepts 0.5 h steps within 0.5–10,000, with a dot or a comma", () => {
		expect(parseHours("12.5")).toBe(12.5);
		expect(parseHours("12,5")).toBe(12.5);
		expect(parseHours("10000")).toBe(10000);
		expect(parseHours("")).toBeNull();
	});

	test("rejects values outside the bounds or off the 0.5 step", () => {
		for (const bad of ["0", "0.25", "10000.5", "12.3", "-1", "abc"]) {
			expect(parseHours(bad)).toBe("invalid");
		}
	});
});

describe("status sheet", () => {
	test("choosing a status queues one PUT and updates the cache optimistically", async () => {
		const sent: QueueEntry[] = [];
		const { queryClient, wrap } = setup(async (entry) => {
			sent.push(entry);
			return { ok: true, body: { status: "completed" } };
		});
		queryClient.setQueryData(queryKeys.itemMe("i1"), {
			rating: null,
			status: null,
			timeToBeat: null,
		});
		const onDone = jest.fn();
		await render(
			wrap(<StatusSheet itemId="i1" type="games" current={null} locale="en" onDone={onDone} />),
		);
		await fireEvent.press(screen.getByRole("radio", { name: "Played" }));
		expect(statusOf(queryClient.getQueryData(queryKeys.itemMe("i1")))).toBe("completed");
		await waitFor(() => expect(sent).toHaveLength(1));
		expect(sent[0]).toMatchObject({
			op: "status",
			itemId: "i1",
			payload: { method: "put", body: { status: "completed" } },
		});
		expect(onDone).toHaveBeenCalled();
	});

	test("labels follow the item type; clearing sends a DELETE", async () => {
		const sent: QueueEntry[] = [];
		const { wrap } = setup(async (entry) => {
			sent.push(entry);
			return { ok: true, body: undefined };
		});
		await render(wrap(<StatusSheet itemId="b1" type="books" current="want" locale="en" />));
		expect(
			screen.getByRole("radio", { name: "Want to read" }).props.accessibilityState,
		).toMatchObject({ checked: true });
		await fireEvent.press(screen.getByRole("button", { name: "No status" }));
		await waitFor(() => expect(sent).toHaveLength(1));
		expect(sent[0]).toMatchObject({ op: "status", payload: { method: "delete" } });
	});
});

describe("v1 look", () => {
	test("status sheet: titled header, the checked option carries a check mark, not colour alone", async () => {
		const { wrap } = setup(ok({}));
		await render(wrap(<StatusSheet itemId="g1" type="games" current="completed" locale="en" />));
		expect(screen.getByRole("header", { name: "Status" })).toBeTruthy();
		const marks = screen.getAllByTestId("status-checked", { includeHiddenElements: true });
		expect(marks).toHaveLength(1);
		expect(screen.getByRole("radio", { name: "Played" })).toContainElement(marks[0] ?? null);
	});

	test("status sheet without a status shows no check mark", async () => {
		const { wrap } = setup(ok({}));
		await render(wrap(<StatusSheet itemId="g1" type="games" current={null} locale="en" />));
		expect(screen.queryAllByTestId("status-checked", { includeHiddenElements: true })).toHaveLength(
			0,
		);
	});

	test("time-to-beat sheet: titled header", async () => {
		const { wrap } = setup(ok({}));
		await render(wrap(<TimeToBeatSheet itemId="g1" current={null} locale="en" />));
		expect(screen.getByRole("header", { name: "Your time to beat" })).toBeTruthy();
	});
});

describe("time-to-beat sheet", () => {
	test("values outside 0.5–10,000 h are rejected client-side and nothing is queued", async () => {
		const send = jest.fn(ok({}));
		const { wrap } = setup(send);
		await render(wrap(<TimeToBeatSheet itemId="g1" current={null} locale="en" />));
		await fireEvent.changeText(screen.getByLabelText("Main story"), "10000.5");
		await fireEvent.press(screen.getByRole("button", { name: "Save times" }));
		expect(screen.getByText("Use 0.5 to 10,000 hours, in steps of 0.5.")).toBeTruthy();
		expect(send).not.toHaveBeenCalled();
	});

	test("a valid report queues one PUT with hours per style", async () => {
		const sent: QueueEntry[] = [];
		const { wrap } = setup(async (entry) => {
			sent.push(entry);
			return { ok: true, body: { timeToBeat: { mainHours: 12.5 } } };
		});
		await render(wrap(<TimeToBeatSheet itemId="g1" current={null} locale="en" />));
		await fireEvent.changeText(screen.getByLabelText("Main story"), "12,5");
		await fireEvent.press(screen.getByRole("button", { name: "Save times" }));
		await waitFor(() => expect(sent).toHaveLength(1));
		expect(sent[0]).toMatchObject({
			op: "timeToBeat",
			itemId: "g1",
			payload: { method: "put", body: { mainHours: 12.5 } },
		});
	});

	test("409 status-not-completed rolls back the cache and shows the errors message", async () => {
		const { queryClient, notify, wrap } = setup(async () => ({
			ok: false,
			status: 409,
			code: "status-not-completed",
			retryAfter: null,
		}));
		const before = { rating: null, status: "in_progress", timeToBeat: { mainHours: 3 } };
		queryClient.setQueryData(queryKeys.itemMe("g1"), before);
		await render(wrap(<TimeToBeatSheet itemId="g1" current={{ mainHours: 3 }} locale="en" />));
		await fireEvent.changeText(screen.getByLabelText("Main story"), "20");
		await fireEvent.press(screen.getByRole("button", { name: "Save times" }));
		await waitFor(() => expect(notify).toHaveBeenCalledWith("errors.status-not-completed"));
		expect(timeToBeatOf(queryClient.getQueryData(queryKeys.itemMe("g1")))).toEqual({
			mainHours: 3,
		});
	});
});

describe("item screen", () => {
	const reader = (type: string): ItemReader => ({
		getItem: jest.fn(async () => ({ id: "x1", type, title: "Title", slug: "title" }) as never),
		listEditions: jest.fn(async () => ({ data: [] }) as never),
		me: jest.fn(async () => ({ rating: null, status: null, timeToBeat: null })),
	});

	test.each([
		["games", true],
		["movies", false],
		["series", false],
		["books", false],
	] as const)("%s: time-to-beat shown = %s, status always", async (type, shown) => {
		const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
		await render(
			<QueryClientProvider client={queryClient}>
				<ItemView type={type} id="x1" read={reader(type)} locale="en" />
			</QueryClientProvider>,
		);
		expect(await screen.findByRole("button", { name: "Status" })).toBeTruthy();
		expect(screen.queryByRole("button", { name: "Your time to beat" }) !== null).toBe(shown);
	});
});
