import type { ReactElement, ReactNode } from "react";
import { Modal, Pressable, Text, View } from "react-native";
import { Button } from "./Button";

export interface SheetProps {
	visible: boolean;
	onClose: () => void;
	title: string;
	closeLabel: string;
	children: ReactNode;
	/** `dialog` centres the panel (confirmations); `sheet` docks it to the bottom edge. */
	variant?: "sheet" | "dialog";
}

// shadcn sheet / dialog: dimmed backdrop that closes on tap, rounded panel, hairline border.
export function Sheet({
	visible,
	onClose,
	title,
	closeLabel,
	children,
	variant = "sheet",
}: SheetProps): ReactElement {
	const dialog = variant === "dialog";
	return (
		<Modal transparent visible={visible} animationType="fade" onRequestClose={onClose}>
			<View className={`flex-1 ${dialog ? "justify-center px-4" : "justify-end"}`}>
				<Pressable
					accessible={false}
					importantForAccessibility="no"
					onPress={onClose}
					className="absolute inset-0 bg-backdrop/50"
				/>
				<View
					accessibilityViewIsModal
					className={`gap-3 border border-border bg-bg p-4 dark:border-border-dark dark:bg-bg-dark ${dialog ? "rounded-lg" : "rounded-t-xl border-b-0"}`}
				>
					<Text
						accessibilityRole="header"
						className="font-display text-20 font-semibold tracking-tight text-fg dark:text-fg-dark"
					>
						{title}
					</Text>
					{children}
					<Button variant="ghost" label={closeLabel} onPress={onClose} />
				</View>
			</View>
		</Modal>
	);
}
