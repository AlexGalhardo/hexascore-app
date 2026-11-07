import type { ReactElement } from "react";
import { Text, View } from "react-native";

export interface ToastProps {
	message: string;
	variant?: "default" | "destructive";
}

// Presentational toast: a polite live region (assertive for errors) on a card, anchored by the caller.
export function Toast({ message, variant = "default" }: ToastProps): ReactElement {
	const destructive = variant === "destructive";
	return (
		<View
			testID="toast"
			accessible
			accessibilityRole="alert"
			accessibilityLiveRegion={destructive ? "assertive" : "polite"}
			className="mx-4 rounded-lg border border-border bg-bg px-4 py-3 dark:border-border-dark dark:bg-bg-dark"
		>
			<Text
				className={`font-text text-14 font-semibold ${destructive ? "text-danger dark:text-danger-dark" : "text-fg dark:text-fg-dark"}`}
			>
				{message}
			</Text>
		</View>
	);
}
