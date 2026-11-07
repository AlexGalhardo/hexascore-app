import { type ReactElement, type ReactNode, useContext } from "react";
import { View } from "react-native";
import { SafeAreaInsetsContext } from "react-native-safe-area-context";

// Edge-to-edge: stack screens without an AppHeader keep their content clear of the status bar.
export function Screen({ children }: { children: ReactNode }): ReactElement {
	// Null outside a provider (unit tests): no inset.
	const top = useContext(SafeAreaInsetsContext)?.top ?? 0;
	return (
		<View style={{ paddingTop: top }} className="flex-1 bg-bg dark:bg-bg-dark">
			{children}
		</View>
	);
}
