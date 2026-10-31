import { type ReactElement, useSyncExternalStore } from "react";
import { Text, View } from "react-native";
import { type AppLocale, t } from "../i18n";
import { getNetworkState, subscribeNetwork } from "./network";

export function OfflineBanner({ locale }: { locale: AppLocale }): ReactElement | null {
	const state = useSyncExternalStore(subscribeNetwork, getNetworkState);
	if (state !== "offline") return null;
	return (
		<View accessibilityLiveRegion="polite" className="bg-surface px-4 py-2 dark:bg-surface-dark">
			<Text className="text-fg dark:text-fg-dark">{t(locale, "account.offline")}</Text>
		</View>
	);
}
