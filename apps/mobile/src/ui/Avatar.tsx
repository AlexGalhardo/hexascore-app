import { Image } from "expo-image";
import type { ReactElement } from "react";
import { StyleSheet, Text, View } from "react-native";

export interface AvatarProps {
	name: string;
	uri?: string | null;
	size?: "sm" | "default" | "lg";
}

const BOX = { sm: "h-8 w-8", default: "h-10 w-10", lg: "h-14 w-14" } as const;
const INITIAL = { sm: "text-12", default: "text-16", lg: "text-24" } as const;

export function initials(name: string): string {
	const parts = name.trim().split(/\s+/).filter(Boolean);
	const first = parts[0] ?? "?";
	const last = parts.length > 1 ? (parts[parts.length - 1] ?? "") : "";
	return `${first.charAt(0)}${last.charAt(0)}`.toUpperCase();
}

export function Avatar({ name, uri = null, size = "default" }: AvatarProps): ReactElement {
	return (
		<View
			accessible
			accessibilityRole="image"
			accessibilityLabel={name}
			className={`${BOX[size]} items-center justify-center overflow-hidden rounded-full bg-hover dark:bg-hover-dark`}
		>
			{uri ? (
				<Image source={uri} contentFit="cover" style={StyleSheet.absoluteFill} />
			) : (
				<Text className={`font-text ${INITIAL[size]} font-semibold text-fg dark:text-fg-dark`}>
					{initials(name)}
				</Text>
			)}
		</View>
	);
}
