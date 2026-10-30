import { Tabs } from "expo-router";
import type { ReactElement } from "react";

export default function TabsLayout(): ReactElement {
	return (
		<Tabs screenOptions={{ headerShown: false }}>
			<Tabs.Screen name="index" options={{ title: "Home" }} />
			<Tabs.Screen name="[type]" options={{ title: "Section" }} />
			<Tabs.Screen name="search" options={{ title: "Search" }} />
			<Tabs.Screen name="lists" options={{ title: "Lists" }} />
			<Tabs.Screen name="profile" options={{ title: "Profile" }} />
		</Tabs>
	);
}
