import { act, render, screen } from "@testing-library/react-native";
import { setNetworkState } from "./network";
import { OfflineBanner } from "./OfflineBanner";

const message = "You are offline. Changes sync when you reconnect.";

test("shows only while the network is offline", async () => {
	setNetworkState("online");
	const view = await render(<OfflineBanner locale="en" />);
	expect(screen.queryByText(message)).toBeNull();
	await act(async () => setNetworkState("offline"));
	expect(screen.getByText(message)).toBeTruthy();
	await act(async () => setNetworkState("unknown"));
	expect(screen.queryByText(message)).toBeNull();
	await view.unmount();
});
