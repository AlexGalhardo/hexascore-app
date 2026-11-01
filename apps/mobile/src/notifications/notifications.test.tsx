import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { fireEvent, render, screen, waitFor } from "@testing-library/react-native";
import { Text } from "react-native";
import { NotificationsScreen } from "./Notifications";
import { type NotificationsApi, useUnread } from "./useUnread";

const rows = (count: number) =>
	Array.from({ length: count }, (_, i) => ({
		id: `n${i}`,
		kind: "report_resolved",
		payload: {},
		readAt: i % 2 === 0 ? null : "2026-10-30T00:00:00Z",
		createdAt: "2026-10-31T12:00:00Z",
	}));

function fakeApi(data = rows(3), unread = 2) {
	const notificationsRead = jest.fn(async () => undefined);
	const api: NotificationsApi = {
		read: {
			notificationsIndex: jest.fn(async () => ({ data, meta: { next_cursor: null, unread } })),
		},
		write: (call) => call({ notificationsRead }),
	};
	return { api, notificationsRead };
}

const client = (): QueryClient =>
	new QueryClient({
		defaultOptions: { queries: { retry: false, gcTime: Number.POSITIVE_INFINITY } },
	});

function Badge({ api }: { api: NotificationsApi }) {
	return <Text testID="badge">{String(useUnread(api))}</Text>;
}

test("the badge reads meta.unread and mark-all clears it with at most 100 ids", async () => {
	const { api, notificationsRead } = fakeApi(rows(240), 120);
	const qc = client();
	await render(
		<QueryClientProvider client={qc}>
			<Badge api={api} />
			<NotificationsScreen api={api} locale="en" />
		</QueryClientProvider>,
	);
	await waitFor(() => expect(screen.getByTestId("badge").props.children).toBe("120"));
	await fireEvent.press(screen.getByRole("button", { name: "Mark all as read" }));
	await waitFor(() => expect(screen.getByTestId("badge").props.children).toBe("0"));
	const sent = (notificationsRead.mock.calls as unknown as [{ body: { ids: string[] } }][]).map(
		([args]) => args.body.ids,
	);
	expect(sent.map((ids) => ids.length)).toEqual([100, 20]);
	expect(sent.flat().every((id) => Number(id.slice(1)) % 2 === 0)).toBe(true);
	expect(api.read.notificationsIndex).toHaveBeenCalledTimes(1);
});

test("opening an unread row marks that one read", async () => {
	const { api, notificationsRead } = fakeApi();
	await render(
		<QueryClientProvider client={client()}>
			<NotificationsScreen api={api} locale="en" />
		</QueryClientProvider>,
	);
	const [first] = await screen.findAllByRole("button", { name: /Your report was reviewed/ });
	await fireEvent.press(first as never);
	expect(notificationsRead).toHaveBeenCalledWith({ body: { ids: ["n0"] } });
});
