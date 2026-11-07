import type { ReactElement, ReactNode } from "react";
import { Text, View } from "react-native";

export function Card({
	children,
	className = "",
}: {
	children: ReactNode;
	className?: string;
}): ReactElement {
	return (
		<View
			className={`gap-3 rounded-lg border border-border bg-surface p-4 dark:border-border-dark dark:bg-surface-dark ${className}`}
		>
			{children}
		</View>
	);
}

export function CardTitle({ text }: { text: string }): ReactElement {
	return (
		<Text
			accessibilityRole="header"
			className="font-display text-18 font-semibold tracking-tight text-fg dark:text-fg-dark"
		>
			{text}
		</Text>
	);
}

export function CardDescription({ text }: { text: string }): ReactElement {
	return <Text className="font-text text-14 text-muted dark:text-muted-dark">{text}</Text>;
}
