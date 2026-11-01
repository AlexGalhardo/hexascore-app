import type { ReactElement, ReactNode } from "react";
import { Pressable, ScrollView, Text, View } from "react-native";
import { GoogleButton, type GoogleButtonProps } from "../GoogleButton";

export interface AuthPageProps {
	kicker: string;
	title: string;
	lede: string;
	google: GoogleButtonProps;
	orLabel: string;
	/** Login: the email form. Signup: the "continue on the web" action (SDD 10: signup opens web /signup). */
	children: ReactNode;
	footer: { prompt: string; action: string; onPress: () => void };
}

const MUTED = "font-text text-16 text-muted dark:text-muted-dark";

// Round 2 login and signup (item 15): Google above the form with an "or" divider.
export function AuthPage(props: AuthPageProps): ReactElement {
	return (
		<ScrollView
			className="flex-1 bg-bg dark:bg-bg-dark"
			contentContainerClassName="gap-6 px-4 pb-8 pt-6"
			keyboardShouldPersistTaps="handled"
		>
			<View className="gap-3 border-b-2 border-fg pb-6 dark:border-fg-dark">
				<Text className="font-text text-12 font-bold uppercase tracking-wider text-primary dark:text-primary-dark">
					{props.kicker}
				</Text>
				<Text
					accessibilityRole="header"
					className="font-display text-36 font-bold text-fg dark:text-fg-dark"
				>
					{props.title}
				</Text>
				<Text className={MUTED}>{props.lede}</Text>
			</View>
			<GoogleButton {...props.google} />
			<View className="flex-row items-center gap-4" accessibilityElementsHidden>
				<View className="h-px flex-1 bg-border dark:bg-border-dark" />
				<Text className="font-text text-14 text-muted dark:text-muted-dark">{props.orLabel}</Text>
				<View className="h-px flex-1 bg-border dark:bg-border-dark" />
			</View>
			{props.children}
			<View className="flex-row flex-wrap items-center gap-x-1">
				<Text className={MUTED}>{props.footer.prompt}</Text>
				<Pressable
					accessibilityRole="link"
					accessibilityLabel={props.footer.action}
					onPress={props.footer.onPress}
					className="h-11 justify-center"
				>
					<Text className="font-text text-16 font-semibold text-primary dark:text-primary-dark">
						{props.footer.action}
					</Text>
				</Pressable>
			</View>
		</ScrollView>
	);
}
