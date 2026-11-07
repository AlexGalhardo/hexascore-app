import type { ReactElement } from "react";
import { Text, View } from "react-native";

export type BadgeVariant = "default" | "secondary" | "outline" | "destructive";

const BOX: Record<BadgeVariant, string> = {
	default: "border-transparent bg-primary dark:bg-primary-dark",
	secondary: "border-transparent bg-hover dark:bg-hover-dark",
	outline: "border-border dark:border-border-dark",
	destructive: "border-transparent bg-danger dark:bg-danger-dark",
};

const TEXT: Record<BadgeVariant, string> = {
	default: "text-on-primary dark:text-on-primary-dark",
	secondary: "text-fg dark:text-fg-dark",
	outline: "text-fg dark:text-fg-dark",
	destructive: "text-on-primary dark:text-on-primary-dark",
};

export function Badge({
	label,
	variant = "default",
}: {
	label: string;
	variant?: BadgeVariant;
}): ReactElement {
	return (
		<View className={`self-start rounded-full border px-2.5 py-0.5 ${BOX[variant]}`}>
			<Text className={`font-text text-12 font-semibold ${TEXT[variant]}`}>{label}</Text>
		</View>
	);
}
