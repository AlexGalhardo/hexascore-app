import { Image } from "expo-image";
import type { ReactElement } from "react";
import { Pressable, Text } from "react-native";
import { useReducedMotion } from "./motion";

export interface GoogleButtonProps {
	/** "Sign in with Google" / "Sign up with Google", translated per Google's wording. */
	label: string;
	scheme: "light" | "dark";
	onPress: () => void;
	disabled?: boolean;
}

// Google sign-in branding colours are fixed by Google, not HexaScore tokens: the one
// third-party exception documented in docs/openpencil/v1-editorial/README.md (round 2, item 15).
const SURFACE = {
	light: "bg-[#FFFFFF] border-[#747775]",
	dark: "bg-[#131314] border-[#8E918F]",
} as const;
const LABEL = { light: "text-[#1F1F1F]", dark: "text-[#E3E3E3]" } as const;

export function GoogleButton({
	label,
	scheme,
	onPress,
	disabled = false,
}: GoogleButtonProps): ReactElement {
	const reduced = useReducedMotion();
	const sink = reduced ? "" : "active:translate-y-px";
	return (
		<Pressable
			accessibilityRole="button"
			accessibilityLabel={label}
			accessibilityState={{ disabled }}
			disabled={disabled}
			onPress={onPress}
			className={`min-h-11 flex-row items-center justify-center gap-3 rounded-sm border border-b-2 px-4 active:border-b ${SURFACE[scheme]} ${sink} ${disabled ? "opacity-50" : ""}`}
		>
			<Image
				testID="google-mark"
				source={require("../../assets/brand/google-g.svg")}
				className="h-5 w-5"
				contentFit="contain"
				accessible={false}
			/>
			<Text className={`font-text text-16 font-semibold ${LABEL[scheme]}`}>{label}</Text>
		</Pressable>
	);
}
