import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { act, fireEvent, render, screen, within } from "@testing-library/react-native";
import { ApiError } from "../../../../packages/app-client/src/api";
import type { NotificationsApi } from "../notifications/useUnread";
import { About } from "./About";
import { AccountExtras } from "./AccountExtras";
import { DeleteAccount } from "./DeleteAccount";
import { ExportData } from "./Export";

const cooldown = (): ApiError =>
	Object.assign(
		new ApiError(429, { status: 429, title: "Cooldown", code: "export-cooldown" } as never),
		{ retryAfter: 2 * 86_400 + 60 },
	);

test("export: 202 confirms; a second export within 7 days shows Retry-After", async () => {
	const meExport = jest
		.fn()
		.mockResolvedValueOnce({ exportId: "e1", status: "queued" })
		.mockRejectedValueOnce(cooldown());
	await render(<ExportData locale="en" write={(call) => call({ meExport })} />);
	await fireEvent.press(screen.getByRole("button", { name: "Request export" }));
	expect(await screen.findByText(/Export requested/)).toBeTruthy();
	await fireEvent.press(screen.getByRole("button", { name: "Request export" }));
	expect(
		await screen.findByText("You can request one export every 7 days. Try again later."),
	).toBeTruthy();
	expect(screen.getByText("You can try again in 3 days.")).toBeTruthy();
});

describe("delete account", () => {
	beforeEach(() => jest.useFakeTimers());
	afterEach(() => jest.useRealTimers());

	test("the confirm button is enabled only after 10 s, then POST /me/deletion", async () => {
		const meDeletionSchedule = jest.fn(async () => ({ deletesAt: "2026-12-01T00:00:00Z" }));
		const onScheduled = jest.fn();
		await render(
			<DeleteAccount
				locale="en"
				write={(call) => call({ meDeletionSchedule })}
				onScheduled={onScheduled}
			/>,
		);
		await fireEvent.press(screen.getByRole("button", { name: "Delete my account" }));
		expect(screen.getByRole("button", { name: "Confirm (10)" })).toBeDisabled();
		await act(async () => jest.advanceTimersByTime(9_000));
		expect(screen.getByRole("button", { name: "Confirm (1)" })).toBeDisabled();
		await act(async () => jest.advanceTimersByTime(1_000));
		const confirm = screen.getByRole("button", { name: "Yes, delete my account" });
		expect(confirm).toBeEnabled();
		await fireEvent.press(confirm);
		expect(meDeletionSchedule).toHaveBeenCalledWith({ body: { confirm: true } });
		expect(onScheduled).toHaveBeenCalledWith("2026-12-01T00:00:00Z");
	});

	test("keep my account closes the dialog and resets the countdown", async () => {
		await render(<DeleteAccount locale="en" write={jest.fn()} onScheduled={() => {}} />);
		await fireEvent.press(screen.getByRole("button", { name: "Delete my account" }));
		await act(async () => jest.advanceTimersByTime(4_000));
		await fireEvent.press(screen.getByRole("button", { name: "Keep my account" }));
		await fireEvent.press(screen.getByRole("button", { name: "Delete my account" }));
		expect(screen.getByRole("button", { name: "Confirm (10)" })).toBeDisabled();
	});
});

describe("v1 look", () => {
	const inbox = (unread: number): NotificationsApi => ({
		read: {
			notificationsIndex: jest.fn(async () => ({ data: [], meta: { unread } })),
		} as never,
		write: jest.fn(),
	});
	const extras = (unread: number) => (
		<QueryClientProvider client={new QueryClient()}>
			<AccountExtras
				locale="en"
				signedIn
				notifications={inbox(unread)}
				write={jest.fn()}
				onScheduled={() => {}}
				onOpen={() => {}}
			/>
		</QueryClientProvider>
	);

	test("notifications row prints the unread count in a badge and in its name", async () => {
		await render(extras(3));
		const row = await screen.findByRole("button", { name: "Notifications, 3 unread" });
		expect(within(row).getByTestId("unread-badge")).toHaveTextContent("3");
	});

	test("no badge without unread notifications", async () => {
		await render(extras(0));
		expect(await screen.findByRole("button", { name: "Notifications" })).toBeTruthy();
		expect(screen.queryByTestId("unread-badge")).toBeNull();
	});

	test("the deletion dialog is a danger panel with a titled header", async () => {
		await render(<DeleteAccount locale="en" write={jest.fn()} onScheduled={() => {}} />);
		await fireEvent.press(screen.getByRole("button", { name: "Delete my account" }));
		const dialog = screen.getByTestId("delete-dialog");
		expect(within(dialog).getByRole("header", { name: "Delete your account?" })).toBeTruthy();
	});

	test("About groups the version and the data attributions under rules", async () => {
		await render(<About locale="en" version="1.2.3" />);
		expect(screen.getByRole("header", { name: "About" })).toBeTruthy();
		const sources = screen.getByTestId("about-sources");
		expect(within(sources).getByTestId("attribution-tmdb")).toBeTruthy();
		expect(within(sources).getByTestId("attribution-igdb")).toBeTruthy();
	});
});
