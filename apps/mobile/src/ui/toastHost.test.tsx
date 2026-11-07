import { act, render, screen } from "@testing-library/react-native";
import { AccessibilityInfo } from "react-native";
import { dismissToast, showToast } from "../rating/toast";
import { ToastHost } from "./ToastHost";

afterEach(async () => {
	await act(async () => dismissToast());
	jest.useRealTimers();
});

test("shows a toast, announces it and hides it after a few seconds", async () => {
	jest.useFakeTimers();
	const announce = jest.spyOn(AccessibilityInfo, "announceForAccessibility").mockImplementation();
	await render(<ToastHost />);
	expect(screen.queryByTestId("toast")).toBeNull();
	await act(async () => showToast("Rating saved"));
	expect(screen.getByText("Rating saved")).toBeTruthy();
	expect(announce).toHaveBeenCalledWith("Rating saved");
	await act(async () => jest.advanceTimersByTime(4100));
	expect(screen.queryByTestId("toast")).toBeNull();
});

test("a newer message replaces the visible one", async () => {
	jest.spyOn(AccessibilityInfo, "announceForAccessibility").mockImplementation();
	await render(<ToastHost />);
	await act(async () => showToast("One"));
	await act(async () => showToast("Two"));
	expect(screen.queryByText("One")).toBeNull();
	expect(screen.getByText("Two")).toBeTruthy();
});
