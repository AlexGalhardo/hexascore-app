import type { ReactElement } from "react";
import { ActivityIndicator, Pressable, Text } from "react-native";

export type ButtonVariant = "default" | "secondary" | "outline" | "ghost" | "destructive";
export type ButtonSize = "sm" | "default" | "lg";

const SURFACE: Record<ButtonVariant, string> = {
	default:
		"bg-primary active:bg-primary-hover dark:bg-primary-dark dark:active:bg-primary-hover-dark",
	secondary: "bg-surface active:bg-hover dark:bg-surface-dark dark:active:bg-hover-dark",
	outline:
		"border border-border bg-bg active:bg-hover dark:border-border-dark dark:bg-bg-dark dark:active:bg-hover-dark",
	ghost: "bg-transparent active:bg-hover dark:active:bg-hover-dark",
	destructive: "bg-danger active:opacity-90 dark:bg-danger-dark",
};

const LABEL: Record<ButtonVariant, string> = {
	default: "text-on-primary dark:text-on-primary-dark",
	secondary: "text-fg dark:text-fg-dark",
	outline: "text-fg dark:text-fg-dark",
	ghost: "text-fg dark:text-fg-dark",
	destructive: "text-on-primary dark:text-on-primary-dark",
};

// Every size keeps a 44 pt target: the small one grows its hit area instead of its box.
const SIZE: Record<ButtonSize, string> = {
	sm: "min-h-9 px-3",
	default: "min-h-11 px-4",
	lg: "min-h-12 px-6",
};

export interface ButtonProps {
	label: string;
	onPress: () => void;
	variant?: ButtonVariant;
	size?: ButtonSize;
	disabled?: boolean;
	loading?: boolean;
	accessibilityHint?: string;
	className?: string;
}

// shadcn/ui button on the app tokens: flat, 8 px radius, hairline border for outline, no ledge.
export function Button({
	label,
	onPress,
	variant = "default",
	size = "default",
	disabled = false,
	loading = false,
	accessibilityHint,
	className = "",
}: ButtonProps): ReactElement {
	const inert = disabled || loading;
	return (
		<Pressable
			accessibilityRole="button"
			accessibilityLabel={label}
			accessibilityHint={accessibilityHint}
			accessibilityState={{ disabled: inert, busy: loading }}
			disabled={inert}
			hitSlop={size === "sm" ? 4 : 0}
			onPress={onPress}
			className={`flex-row items-center justify-center gap-2 rounded-md ${SIZE[size]} ${SURFACE[variant]} ${disabled ? "opacity-50" : ""} ${className}`}
		>
			{loading ? <ActivityIndicator size="small" /> : null}
			<Text
				className={`font-text font-semibold ${size === "sm" ? "text-14" : "text-16"} ${LABEL[variant]}`}
			>
				{label}
			</Text>
		</Pressable>
	);
}
