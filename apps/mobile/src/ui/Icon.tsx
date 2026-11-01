import Ionicons from "@expo/vector-icons/Ionicons";
import { cssInterop } from "nativewind";
import type { ComponentProps, ReactElement } from "react";

// Lets token classes (text-fg dark:text-fg-dark) colour the glyph.
cssInterop(Ionicons, { className: { target: "style", nativeStyleToProp: { color: true } } });

export type IconName = ComponentProps<typeof Ionicons>["name"];

export interface IconProps {
	name: IconName;
	size: number;
	className?: string;
	color?: string;
}

/** Decorative glyph: the control around it carries the accessible name. */
export function Icon({ name, size, className, color }: IconProps): ReactElement {
	return (
		<Ionicons
			testID={`icon-${name}`}
			name={name}
			size={size}
			color={color}
			className={className}
			accessible={false}
			importantForAccessibility="no-hide-descendants"
		/>
	);
}
