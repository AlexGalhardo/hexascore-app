import type { ReactElement } from "react";
import { Pressable, ScrollView, Text, View } from "react-native";
import { Btn, type BtnVariant } from "../Btn";

export interface StateAction {
	label: string;
	onPress: () => void;
	variant?: BtnVariant;
	/** `link` for actions that leave the app (store, web). */
	role?: "button" | "link";
}

export interface StatePageProps {
	/** Big display number (404, 500); omitted for empty and blocked states. */
	code?: string;
	title: string;
	body: string[];
	actions: StateAction[];
}

const FG = "text-fg dark:text-fg-dark";

function Action({ action }: { action: StateAction }): ReactElement {
	if (action.role !== "link") {
		return <Btn label={action.label} onPress={action.onPress} variant={action.variant} />;
	}
	return (
		<Pressable
			accessibilityRole="link"
			accessibilityLabel={action.label}
			onPress={action.onPress}
			className="min-h-11 items-center justify-center px-4"
		>
			<Text className="font-text text-16 font-semibold text-primary underline dark:text-primary-dark">
				{action.label}
			</Text>
		</Pressable>
	);
}

// v1 web-mobile error/404 and error/500: display code, 2 px rule, serif title, muted body, stacked actions.
export function StatePage({ code, title, body, actions }: StatePageProps): ReactElement {
	return (
		<ScrollView
			className="flex-1 bg-bg dark:bg-bg-dark"
			contentContainerClassName="grow items-center justify-center gap-4 px-4 py-12"
		>
			<View testID="state-page" accessibilityLiveRegion="polite" className="items-center gap-4">
				{code ? (
					// The title says the same thing in words; reading the number too is noise.
					<View accessibilityElementsHidden importantForAccessibility="no-hide-descendants">
						<Text className="font-display text-36 font-bold text-primary dark:text-primary-dark">
							{code}
						</Text>
						<View className="h-0.5 w-32 self-center bg-fg dark:bg-fg-dark" />
					</View>
				) : null}
				<Text
					accessibilityRole="header"
					className={`text-center font-display text-30 font-bold ${FG}`}
				>
					{title}
				</Text>
				{body.map((line) => (
					<Text
						key={line}
						className="max-w-xs text-center font-text text-16 leading-relaxed text-muted dark:text-muted-dark"
					>
						{line}
					</Text>
				))}
				{actions.length > 0 ? (
					<View className="w-full max-w-xs gap-3 pt-2">
						{actions.map((action) => (
							<Action key={action.label} action={action} />
						))}
					</View>
				) : null}
			</View>
		</ScrollView>
	);
}
