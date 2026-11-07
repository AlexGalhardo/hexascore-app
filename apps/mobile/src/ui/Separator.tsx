import type { ReactElement } from "react";
import { View } from "react-native";

export function Separator({ vertical = false }: { vertical?: boolean }): ReactElement {
	return (
		<View
			accessibilityElementsHidden
			importantForAccessibility="no-hide-descendants"
			className={`bg-border dark:bg-border-dark ${vertical ? "w-px self-stretch" : "h-px w-full"}`}
		/>
	);
}
