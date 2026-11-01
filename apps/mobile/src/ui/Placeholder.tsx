import type { ReactElement } from "react";
import { Text, View } from "react-native";

export function Placeholder({ title }: { title: string }): ReactElement {
	return (
		<View className="flex-1 items-center justify-center bg-bg px-4 dark:bg-bg-dark">
			<Text className="font-display text-20 font-bold text-fg dark:text-fg-dark">{title}</Text>
		</View>
	);
}
