import type { ReactElement, ReactNode } from "react";
import { ScrollView, Text, View } from "react-native";
import type { ThemePreference } from "../../theme";
import { Button } from "../Button";
import { SegmentedControl } from "../SegmentedControl";

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
	/** Signed-in sections (profile form, devices) between Account and Appearance. */
	children?: ReactNode;
	onTheme: (theme: ThemePreference) => void;
	onSignOut: () => void;
	onSignIn: () => void;
}

const MODES: readonly ThemePreference[] = ["system", "light", "dark"];
const FG = "text-fg dark:text-fg-dark";
const SECTION = "gap-3 border-t border-border pt-4 dark:border-border-dark";

function Heading({ children }: { children: string }): ReactElement {
	return (
		<Text
			accessibilityRole="header"
			className={`font-display text-24 font-semibold tracking-tight ${FG}`}
		>
			{children}
		</Text>
	);
}

/** A titled settings section, for the sections passed as children. */
export function SettingsSection({
	title,
	children,
}: {
	title: string;
	children: ReactNode;
}): ReactElement {
	return (
		<View className={SECTION}>
			<Heading>{title}</Heading>
			{children}
		</View>
	);
}

export function SettingsPage(props: SettingsPageProps): ReactElement {
	const { labels } = props;
	return (
		<ScrollView
			className="flex-1 bg-bg dark:bg-bg-dark"
			contentContainerClassName="gap-6 px-4 pb-8 pt-6"
		>
			<Text
				accessibilityRole="header"
				className={`font-display text-36 font-semibold tracking-tight ${FG}`}
			>
				{props.title}
			</Text>
			<View className={SECTION}>
				<Heading>{labels.account}</Heading>
				{props.account === null ? (
					<Button label={labels.signIn} onPress={props.onSignIn} />
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
						<Button variant="outline" label={labels.signOut} onPress={props.onSignOut} />
					</>
				)}
			</View>
			{props.children}
			<View className={SECTION}>
				<Heading>{labels.appearance}</Heading>
				<SegmentedControl
					radio
					label={labels.appearance}
					value={props.theme}
					onChange={props.onTheme}
					options={MODES.map((mode) => ({ value: mode, label: labels.theme[mode] }))}
				/>
			</View>
			<View className={SECTION}>
				<Heading>{labels.privacy}</Heading>
				{props.crashReport}
			</View>
		</ScrollView>
	);
}
