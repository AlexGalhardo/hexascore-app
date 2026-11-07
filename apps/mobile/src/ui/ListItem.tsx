import type { ReactElement, ReactNode } from "react";
import { Pressable, Text, View } from "react-native";

export interface ListItemProps {
	title: string;
	subtitle?: string;
	leading?: ReactNode;
	trailing?: ReactNode;
	onPress?: () => void;
	accessibilityLabel?: string;
}

const BOX =
	"min-h-11 flex-row items-center gap-3 border-b border-border py-3 dark:border-border-dark";

// A row closed by a hairline; a 44 pt button when onPress is given.
export function ListItem({
	title,
	subtitle,
	leading,
	trailing,
	onPress,
	accessibilityLabel,
}: ListItemProps): ReactElement {
	const body = (
		<>
			{leading}
			<View className="flex-1">
				<Text className="font-text text-16 font-semibold text-fg dark:text-fg-dark">{title}</Text>
				{subtitle ? (
					<Text className="font-text text-14 text-muted dark:text-muted-dark">{subtitle}</Text>
				) : null}
			</View>
			{trailing}
		</>
	);
	if (!onPress) return <View className={BOX}>{body}</View>;
	return (
		<Pressable
			accessibilityRole="button"
			accessibilityLabel={accessibilityLabel ?? [title, subtitle].filter(Boolean).join(", ")}
			onPress={onPress}
			className={`${BOX} active:bg-surface dark:active:bg-surface-dark`}
		>
			{body}
		</Pressable>
	);
}
