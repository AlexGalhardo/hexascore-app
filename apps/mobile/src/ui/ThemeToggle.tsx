import type { ReactElement } from "react";
import { Pressable, Text } from "react-native";

export interface ThemeToggleProps {
	scheme: "light" | "dark";
	onChange: (next: "light" | "dark") => void;
	labels: { toDark: string; toLight: string };
}

// Round 2 item 8: moon in light, sun in dark.
export function ThemeToggle({ scheme, onChange, labels }: ThemeToggleProps): ReactElement {
	const next = scheme === "light" ? "dark" : "light";
	return (
		<Pressable
			accessibilityRole="button"
			accessibilityLabel={next === "dark" ? labels.toDark : labels.toLight}
			onPress={() => onChange(next)}
			className="h-11 w-11 items-center justify-center rounded-full"
		>
			<Text className="text-20 text-fg dark:text-fg-dark">{scheme === "light" ? "☾" : "☀"}</Text>
		</Pressable>
	);
}
