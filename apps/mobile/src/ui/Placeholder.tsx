import type { ReactElement } from "react";
import { View } from "react-native";
import { EmptyState } from "./EmptyState";

export function Placeholder({ title }: { title: string }): ReactElement {
	return (
		<View className="flex-1 justify-center bg-bg px-4 dark:bg-bg-dark">
			<EmptyState title={title} />
		</View>
	);
}
