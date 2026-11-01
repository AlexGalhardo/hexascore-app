import type { ReactElement, ReactNode } from "react";
import { Pressable, ScrollView, Text, View } from "react-native";
import type { ThemePreference } from "../../theme";
import { Btn } from "../Btn";

export interface SettingsLabels {
	account: string;
	appearance: string;
	theme: Record<ThemePreference, string>;
	privacy: string;
	signOut: string;
	signIn: string;
}

export interface SettingsPageProps {
	title: string;
	/** Null when signed out; a field is null until `GET /me` has loaded. */
	account: { name: string | null; email: string | null } | null;
	theme: ThemePreference;
	/** Crash-report consent switch, wired from src/telemetry. */
	crashReport: ReactNode;
	labels: SettingsLabels;
	onTheme: (theme: ThemePreference) => void;
	onSignOut: () => void;
	onSignIn: () => void;
}

const MODES: readonly ThemePreference[] = ["system", "light", "dark"];
const FG = "text-fg dark:text-fg-dark";
const SECTION = "gap-3 border-t-2 border-fg pt-4 dark:border-fg-dark";

function Heading({ children }: { children: string }): ReactElement {
	return (
		<Text accessibilityRole="header" className={`font-display text-24 font-bold ${FG}`}>
			{children}
		</Text>
	);
}

export function SettingsPage(props: SettingsPageProps): ReactElement {
	const { labels } = props;
	return (
		<ScrollView
			className="flex-1 bg-bg dark:bg-bg-dark"
			contentContainerClassName="gap-6 px-4 pb-8 pt-6"
		>
			<Text accessibilityRole="header" className={`font-display text-36 font-bold ${FG}`}>
				{props.title}
			</Text>
			<View className={SECTION}>
				<Heading>{labels.account}</Heading>
				{props.account === null ? (
					<Btn label={labels.signIn} onPress={props.onSignIn} />
				) : (
					<>
						<View>
							{props.account.name ? (
								<Text className={`font-text text-16 font-semibold ${FG}`}>
									{props.account.name}
								</Text>
							) : null}
							{props.account.email ? (
								<Text className="font-text text-14 text-muted dark:text-muted-dark">
									{props.account.email}
								</Text>
							) : null}
						</View>
						<Btn variant="secondary" label={labels.signOut} onPress={props.onSignOut} />
					</>
				)}
			</View>
			<View className={SECTION}>
				<Heading>{labels.appearance}</Heading>
				<View accessibilityRole="radiogroup" className="flex-row">
					{MODES.map((mode) => {
						const on = props.theme === mode;
						return (
							<Pressable
								key={mode}
								accessibilityRole="radio"
								accessibilityLabel={labels.theme[mode]}
								accessibilityState={{ checked: on }}
								onPress={() => props.onTheme(mode)}
								className={`min-h-11 flex-1 items-center justify-center border border-fg dark:border-fg-dark ${on ? "bg-fg dark:bg-fg-dark" : "bg-field-bg dark:bg-field-bg-dark"}`}
							>
								<Text
									className={`font-text text-16 font-semibold ${on ? "text-bg dark:text-bg-dark" : FG}`}
								>
									{labels.theme[mode]}
								</Text>
							</Pressable>
						);
					})}
				</View>
			</View>
			<View className={SECTION}>
				<Heading>{labels.privacy}</Heading>
				{props.crashReport}
			</View>
		</ScrollView>
	);
}
