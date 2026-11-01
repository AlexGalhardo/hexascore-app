import { Tabs, useRouter } from "expo-router";
import { useColorScheme } from "nativewind";
import { type ReactElement, useEffect, useSyncExternalStore } from "react";
import { useSafeAreaInsets } from "react-native-safe-area-context";
import { session } from "../../src/auth/runtime";
import { t } from "../../src/i18n";
import { notificationsApi } from "../../src/notifications/runtime";
import { useUnread } from "../../src/notifications/useUnread";
import { memoryPreferences } from "../../src/platform/preferences";
import { useLocale } from "../../src/providers/root";
import { applyThemePreference, restoreThemePreference } from "../../src/theme";
import { AppHeader } from "../../src/ui/AppHeader";
import { TabIcon, type TabRoute } from "../../src/ui/TabIcon";
import { profileBadge, tabBarTheme } from "../../src/ui/tabBarTheme";

// Round 2 order: Home, Search, Sections, Lists, Profile.
const TABS: readonly [TabRoute, string][] = [
	["index", "nav.home"],
	["search", "nav.searchTab"],
	["[type]", "nav.sections"],
	["lists", "nav.lists"],
	["profile", "nav.profile"],
];

export default function TabsLayout(): ReactElement {
	const locale = useLocale();
	const router = useRouter();
	const { top } = useSafeAreaInsets();
	const scheme = useColorScheme().colorScheme === "dark" ? "dark" : "light";
	const signedIn = useSyncExternalStore(session.subscribe, session.state).kind === "active";
	const badge = profileBadge(useUnread(notificationsApi, signedIn));
	useEffect(() => {
		restoreThemePreference(memoryPreferences);
	}, []);
	return (
		<Tabs
			screenOptions={{
				...tabBarTheme(scheme),
				header: () => (
					<AppHeader
						scheme={scheme}
						topInset={top}
						labels={{
							search: t(locale, "nav.openSearch"),
							toDark: t(locale, "nav.themeDark"),
							toLight: t(locale, "nav.themeLight"),
						}}
						onSearch={() => router.navigate("/search")}
						onToggleTheme={(next) => applyThemePreference(memoryPreferences, next)}
					/>
				),
			}}
		>
			{TABS.map(([route, key]) => (
				<Tabs.Screen
					key={route}
					name={route}
					options={{
						title: t(locale, key),
						tabBarIcon: (props) => <TabIcon route={route} {...props} />,
						tabBarBadge: route === "profile" ? badge : undefined,
					}}
				/>
			))}
		</Tabs>
	);
}
