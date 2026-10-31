import "../global.css";
import { PersistQueryClientProvider } from "@tanstack/react-query-persist-client";
import { Stack } from "expo-router";
import type { ReactElement } from "react";
import { appLocale, persistOptions, queryClient, rootDeps } from "../src/auth/runtime";
import { LocaleProvider, routeGuards, useRoot } from "../src/providers/root";

export default function RootLayout(): ReactElement | null {
	const { state, upgrade } = useRoot(rootDeps);
	// Every guard is false until the vault is read; render nothing instead of an unmatched route.
	if (state.kind === "loading" && !upgrade) return null;
	const guards = routeGuards(state, upgrade);
	return (
		<PersistQueryClientProvider client={queryClient} persistOptions={persistOptions}>
			<LocaleProvider locale={appLocale}>
				<Stack screenOptions={{ headerShown: false }}>
					<Stack.Protected guard={guards.update}>
						<Stack.Screen name="update-required" />
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
