import type { ReactElement } from "react";
import { Text } from "react-native";

export function Label({ text }: { text: string }): ReactElement {
	return <Text className="font-text text-14 font-semibold text-fg dark:text-fg-dark">{text}</Text>;
}
