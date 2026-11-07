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

// Placeholder rows shaped like ItemRow; one polite, labelled region for screen readers.
export function ListSkeleton({ label, rows = 4 }: { label: string; rows?: number }): ReactElement {
	return (
		<View accessible accessibilityLabel={label} accessibilityLiveRegion="polite" className="gap-4">
			{Array.from({ length: rows }, (_, index) => (
				// biome-ignore lint/suspicious/noArrayIndexKey: static placeholder rows
				<View key={index} className="flex-row items-center gap-4">
					<Skeleton className="h-16 w-10" />
					<View className="flex-1 gap-2">
						<Skeleton className="h-4 w-3/4" />
						<Skeleton className="h-3 w-1/2" />
					</View>
					<Skeleton className="h-9 w-11" />
				</View>
			))}
		</View>
	);
}
