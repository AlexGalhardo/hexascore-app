import { type ReactElement, useState } from "react";
import { Switch, Text, View } from "react-native";
import { type AppLocale, t } from "../i18n";
import { crashConsent, setCrashConsent } from "./runtime";

/** Revocable crash-report consent (P5.A.4.3); "unset" reads as off. */
export function CrashReportSwitch({ locale }: { locale: AppLocale }): ReactElement {
	const [on, setOn] = useState(() => crashConsent() === "granted");
	const label = t(locale, "account.settings.crashReports");
	return (
		<View className="min-h-11 flex-row items-center gap-4">
			<View className="flex-1 gap-1">
				<Text className="font-text text-16 font-semibold text-fg dark:text-fg-dark">{label}</Text>
				<Text className="font-text text-14 text-muted dark:text-muted-dark">
					{t(locale, "account.crashReports.body")}
				</Text>
			</View>
			<Switch
				accessibilityRole="switch"
				accessibilityLabel={label}
				value={on}
				onValueChange={(next) => {
					setOn(next);
					void setCrashConsent(next);
				}}
			/>
		</View>
	);
}
