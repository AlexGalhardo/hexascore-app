import { type ReactElement, useEffect, useSyncExternalStore } from "react";
import { AccessibilityInfo, View } from "react-native";
import { currentToast, dismissToast, subscribeToast } from "../rating/toast";
import { Toast } from "./Toast";

const VISIBLE_MS = 4000;

// Mounted once at the root: shows the latest showToast message above the tab bar and announces it.
export function ToastHost(): ReactElement | null {
	const toast = useSyncExternalStore(subscribeToast, currentToast, currentToast);
	useEffect(() => {
		if (toast === null) return;
		AccessibilityInfo.announceForAccessibility(toast.message);
		const timer = setTimeout(dismissToast, VISIBLE_MS);
		return () => clearTimeout(timer);
	}, [toast]);
	if (toast === null) return null;
	return (
		<View pointerEvents="none" className="absolute inset-x-0 bottom-24">
			<Toast message={toast.message} />
		</View>
	);
}
