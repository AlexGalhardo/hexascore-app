import "../global.css";
import { Stack } from "expo-router";
import type { ReactElement } from "react";
import { appLocale, rootDeps } from "../src/auth/runtime";
import { LocaleProvider, routeGuards, useRoot } from "../src/providers/root";

export default function RootLayout(): ReactElement {
	const { state, upgrade } = useRoot(rootDeps);
	const guards = routeGuards(state, upgrade);
	return (
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
	);
}
