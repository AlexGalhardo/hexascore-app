import type { ReactElement } from "react";
import { Pressable, Text, View } from "react-native";
import { type AppLocale, t } from "../i18n";

export function CrashConsentPrompt({
	locale,
	onChoose,
}: {
	locale: AppLocale;
	onChoose: (granted: boolean) => void;
}): ReactElement {
	return (
		<View accessibilityRole="alert" className="gap-2 bg-surface px-4 py-4 dark:bg-surface-dark">
			<Text className="text-fg dark:text-fg-dark">{t(locale, "account.crashReports.title")}</Text>
			<Text className="text-fg dark:text-fg-dark">{t(locale, "account.crashReports.body")}</Text>
			<View className="flex-row gap-4">
				<Pressable
					accessibilityRole="button"
					className="min-h-11 justify-center"
					onPress={() => onChoose(false)}
				>
					<Text className="text-fg dark:text-fg-dark">
						{t(locale, "account.crashReports.deny")}
					</Text>
				</Pressable>
				<Pressable
					accessibilityRole="button"
					className="min-h-11 justify-center"
					onPress={() => onChoose(true)}
				>
					<Text className="text-fg dark:text-fg-dark">
						{t(locale, "account.crashReports.allow")}
					</Text>
				</Pressable>
			</View>
		</View>
	);
}
