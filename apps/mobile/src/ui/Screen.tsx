import type { ReactElement, ReactNode } from "react";
import { View } from "react-native";
import { useSafeAreaInsets } from "react-native-safe-area-context";

// Edge-to-edge: stack screens without an AppHeader keep their content clear of the status bar.
export function Screen({ children }: { children: ReactNode }): ReactElement {
	const { top } = useSafeAreaInsets();
	return (
		<View style={{ paddingTop: top }} className="flex-1 bg-bg dark:bg-bg-dark">
			{children}
		</View>
	);
}
