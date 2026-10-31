import "../global.css";
import { PersistQueryClientProvider } from "@tanstack/react-query-persist-client";
import { type ErrorBoundaryProps, Stack, useSegments } from "expo-router";
import { type ReactElement, useEffect, useState } from "react";
import { Pressable, Text, View } from "react-native";
import { appLocale, persistOptions, queryClient, rootDeps } from "../src/auth/runtime";
import { t } from "../src/i18n";
import { OfflineBanner } from "../src/platform/OfflineBanner";
import { LocaleProvider, routeGuards, useRoot } from "../src/providers/root";
import { CrashConsentPrompt } from "../src/telemetry/CrashConsentPrompt";
import {
	crashConsent,
	reportCrash,
	setCrashConsent,
	setCrashScreen,
} from "../src/telemetry/runtime";

// Expo Router renders this when a route throws during render (the error boundary of SDD 10 §7.7).
export function ErrorBoundary({ error, retry }: ErrorBoundaryProps): ReactElement {
	useEffect(() => reportCrash(error), [error]);
	return (
		<View className="flex-1 items-center justify-center gap-4 bg-bg px-4 dark:bg-bg-dark">
			<Text className="text-20 text-fg dark:text-fg-dark">
				{t(appLocale, "errors.serverTitle")}
			</Text>
			<Pressable
				accessibilityRole="button"
				className="min-h-11 justify-center"
				onPress={() => void retry()}
			>
				<Text className="text-fg dark:text-fg-dark">{t(appLocale, "errors.retry")}</Text>
			</Pressable>
		</View>
	);
}

export default function RootLayout(): ReactElement | null {
	const { state, upgrade } = useRoot(rootDeps);
	// Route pattern segments (`[type]`, `[slug]`), never the concrete values, so no username reaches the report.
	// The typed-routes manifest is not generated in this repo, so the hook is typed `never`.
	const segments = (useSegments as () => string[])();
	const [consent, setConsent] = useState(crashConsent);
	useEffect(() => setCrashScreen(`/${segments.join("/")}`), [segments]);
	// Every guard is false until the vault is read; render nothing instead of an unmatched route.
	if (state.kind === "loading" && !upgrade) return null;
	const guards = routeGuards(state, upgrade);
	return (
		<PersistQueryClientProvider client={queryClient} persistOptions={persistOptions}>
			<LocaleProvider locale={appLocale}>
				<OfflineBanner locale={appLocale} />
				{consent === "unset" ? (
					<CrashConsentPrompt
						locale={appLocale}
						onChoose={(granted) => {
							setConsent(granted ? "granted" : "denied");
							void setCrashConsent(granted);
						}}
					/>
				) : null}
				<Stack screenOptions={{ headerShown: false }}>
					<Stack.Protected guard={guards.update}>
						<Stack.Screen name="update-required" />
					</Stack.Protected>
					<Stack.Protected guard={guards.banned}>
						<Stack.Screen name="banned" />
					</Stack.Protected>
					<Stack.Protected guard={guards.deletion}>
						<Stack.Screen name="deletion-scheduled" />
					</Stack.Protected>
					<Stack.Protected guard={guards.auth}>
						<Stack.Screen name="(auth)/login" />
					</Stack.Protected>
					<Stack.Protected guard={guards.app}>
						<Stack.Screen name="(tabs)" />
						<Stack.Screen name="[type]/[slug]" />
						<Stack.Screen name="entities/[id]" />
						<Stack.Screen name="lists/[id]" />
					</Stack.Protected>
				</Stack>
			</LocaleProvider>
		</PersistQueryClientProvider>
	);
}
