import type { ReactElement } from "react";
import { Pressable, Text } from "react-native";
import { useReducedMotion } from "./motion";

export type BtnVariant = "primary" | "secondary" | "ghost";

const SURFACE: Record<BtnVariant, string> = {
	primary:
		"bg-primary dark:bg-primary-dark border-ledge-primary dark:border-ledge-primary-dark active:bg-primary-hover dark:active:bg-primary-hover-dark",
	secondary:
		"bg-field-bg dark:bg-field-bg-dark border border-fg dark:border-fg-dark border-b-ledge-secondary dark:border-b-ledge-secondary-dark",
	ghost: "bg-transparent border-transparent",
};

const LABEL: Record<BtnVariant, string> = {
	primary: "text-on-primary dark:text-on-primary-dark",
	secondary: "text-fg dark:text-fg-dark",
	ghost: "text-primary dark:text-primary-dark underline",
};

export interface BtnProps {
	label: string;
	onPress: () => void;
	variant?: BtnVariant;
	disabled?: boolean;
	accessibilityHint?: string;
}

// v1 round 2 item 12: 2 px ledge at rest; pressed sinks 1 px onto a 1 px ledge.
// Reduced motion keeps the ledge change and drops the movement.
export function Btn({
	label,
	onPress,
	variant = "primary",
	disabled = false,
	accessibilityHint,
}: BtnProps): ReactElement {
	const reduced = useReducedMotion();
	const ledge = variant === "ghost" ? "" : "border-b-2 active:border-b";
	const sink = variant === "ghost" || reduced ? "" : "active:translate-y-px";
	return (
		<Pressable
			accessibilityRole="button"
			accessibilityLabel={label}
			accessibilityHint={accessibilityHint}
			accessibilityState={{ disabled }}
			disabled={disabled}
			onPress={onPress}
			className={`min-h-11 flex-row items-center justify-center rounded-sm px-4 ${SURFACE[variant]} ${ledge} ${sink} ${disabled ? "opacity-50" : ""}`}
		>
			<Text className={`text-16 font-semibold ${LABEL[variant]}`}>{label}</Text>
		</Pressable>
	);
}
