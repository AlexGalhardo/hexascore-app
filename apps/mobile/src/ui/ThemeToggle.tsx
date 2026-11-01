import type { ReactElement } from "react";
import { Pressable } from "react-native";
import { Icon } from "./Icon";

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
			<Icon
				name={scheme === "light" ? "moon-outline" : "sunny-outline"}
				size={22}
				className="text-fg dark:text-fg-dark"
			/>
		</Pressable>
	);
}
