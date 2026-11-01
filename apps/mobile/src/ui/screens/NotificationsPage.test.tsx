import { fireEvent, render, screen } from "@testing-library/react-native";
import { NotificationsPage, type NotificationsProps } from "./NotificationsPage";

function setup(overrides: Partial<NotificationsProps> = {}): NotificationsProps {
	return {
		title: "Notifications",
		rows: [
			{ id: "a", text: "Your report was reviewed.", time: "2 h ago", unread: true },
			{ id: "b", text: "Your export is ready.", time: "Yesterday", unread: false },
		],
		labels: {
			unreadCount: (n) => `${n} unread`,
			markAll: "Mark all as read",
			unread: "Unread",
			empty: "No notifications yet.",
		},
		onOpen: jest.fn(),
		onMarkAll: jest.fn(),
		...overrides,
	};
}

describe("NotificationsPage", () => {
	test("unread rows say so in text, not only colour; mark all and open", async () => {
		const props = setup();
		await render(<NotificationsPage {...props} />);
		expect(screen.getByRole("header", { name: "Notifications" })).toBeTruthy();
		expect(screen.getByText("1 unread")).toBeTruthy();
		expect(
			screen.getByRole("button", { name: "Unread, Your report was reviewed., 2 h ago" }),
		).toBeTruthy();
		await fireEvent.press(screen.getByRole("button", { name: "Your export is ready., Yesterday" }));
		expect(props.onOpen).toHaveBeenCalledWith(expect.objectContaining({ id: "b" }));
		await fireEvent.press(screen.getByRole("button", { name: "Mark all as read" }));
		expect(props.onMarkAll).toHaveBeenCalledTimes(1);
	});

	test("nothing unread disables mark all; no rows shows the empty text", async () => {
		await render(<NotificationsPage {...setup({ rows: [] })} />);
		expect(
			screen.getByRole("button", { name: "Mark all as read" }).props.accessibilityState,
		).toMatchObject({ disabled: true });
		expect(screen.getByText("No notifications yet.")).toBeTruthy();
	});
});
