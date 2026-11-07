import type { ReactElement } from "react";
import { Text, View } from "react-native";
import { Button } from "./Button";

export interface EmptyStateProps {
	title: string;
	description?: string;
	action?: { label: string; onPress: () => void };
}

export function EmptyState({ title, description, action }: EmptyStateProps): ReactElement {
	return (
		<View
			testID="empty-state"
			accessibilityLiveRegion="polite"
			className="items-center gap-2 rounded-lg border border-dashed border-border px-4 py-8 dark:border-border-dark"
		>
			<Text
				accessibilityRole="header"
				className="text-center font-display text-18 font-semibold text-fg dark:text-fg-dark"
			>
				{title}
			</Text>
			{description ? (
				<Text className="text-center font-text text-14 text-muted dark:text-muted-dark">
					{description}
				</Text>
			) : null}
			{action ? (
				<View className="pt-2">
					<Button variant="outline" label={action.label} onPress={action.onPress} />
				</View>
			) : null}
		</View>
	);
}
