import { focusManager, onlineManager, QueryClient } from "@tanstack/react-query";
import { subscribeAppActive } from "../platform/appState";
import { subscribeNetwork } from "../platform/network";

// Legacy web defaults (SDD 10 §7.1).
export function createQueryClient(): QueryClient {
	return new QueryClient({ defaultOptions: { queries: { staleTime: 30_000, retry: 1 } } });
}

/** Routes focus and online state through src/platform/; returns the unwire function. */
export function wireQueryManagers(): () => void {
	focusManager.setEventListener((setFocused) => subscribeAppActive(setFocused));
	// "unknown" counts as online so the first request is not held back before reachability reports.
	onlineManager.setEventListener((setOnline) =>
		subscribeNetwork((state) => setOnline(state !== "offline")),
	);
	return () => {
		focusManager.setEventListener(() => undefined);
		onlineManager.setEventListener(() => undefined);
	};
}
