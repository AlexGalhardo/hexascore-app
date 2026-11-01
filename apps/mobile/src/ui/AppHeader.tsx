import type { ReactElement } from "react";
import { Pressable, Text, View } from "react-native";
import { Icon } from "./Icon";
import { ThemeToggle } from "./ThemeToggle";

export interface AppHeaderProps {
	scheme: "light" | "dark";
	/** Status bar height from `useSafeAreaInsets().top`. */
	topInset: number;
	labels: { search: string; toDark: string; toLight: string };
	onSearch: () => void;
	onToggleTheme: (next: "light" | "dark") => void;
}

// v1 android home: serif wordmark left; search (round 2 item 16: icon only on mobile) and theme toggle right.
export function AppHeader({
	scheme,
	topInset,
	labels,
	onSearch,
	onToggleTheme,
}: AppHeaderProps): ReactElement {
	return (
		<View
			testID="app-header"
			style={{ paddingTop: topInset }}
			className="flex-row items-center justify-between bg-bg px-4 pb-2 dark:bg-bg-dark"
		>
			<Text
				accessibilityRole="header"
				className="font-display text-24 font-bold text-fg dark:text-fg-dark"
			>
				HexaScore
			</Text>
			<View className="flex-row items-center">
				<Pressable
					accessibilityRole="button"
					accessibilityLabel={labels.search}
					onPress={onSearch}
					className="h-11 w-11 items-center justify-center rounded-full"
				>
					<Icon name="search-outline" size={22} className="text-fg dark:text-fg-dark" />
				</Pressable>
				<ThemeToggle
					scheme={scheme}
					onChange={onToggleTheme}
					labels={{ toDark: labels.toDark, toLight: labels.toLight }}
				/>
			</View>
		</View>
	);
}
