import type { ReactElement } from "react";
import { View } from "react-native";

// A static block, no shimmer: reduced motion is honoured by construction.
export function Skeleton({ className = "h-4 w-full" }: { className?: string }): ReactElement {
	return (
		<View
			testID="skeleton"
			accessibilityElementsHidden
			importantForAccessibility="no-hide-descendants"
			className={`rounded-md bg-hover dark:bg-hover-dark ${className}`}
		/>
	);
}
