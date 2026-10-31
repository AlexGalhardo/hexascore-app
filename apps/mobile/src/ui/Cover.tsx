import { Image } from "expo-image";
import type { ReactElement } from "react";
import { StyleSheet, View } from "react-native";

export interface CoverProps {
	uri: string | null;
	title: string;
	className?: string;
}

// Reserved 2:3 box so the layout does not shift while the image loads.
export function Cover({ uri, title, className = "" }: CoverProps): ReactElement {
	return (
		<View
			testID="cover"
			accessible
			accessibilityRole="image"
			accessibilityLabel={title}
			className={`aspect-[2/3] overflow-hidden rounded-sm bg-surface dark:bg-surface-dark ${className}`}
		>
			{uri !== null && (
				<Image source={uri} contentFit="cover" transition={0} style={StyleSheet.absoluteFill} />
			)}
		</View>
	);
}
