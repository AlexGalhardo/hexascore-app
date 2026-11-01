import { fontFamily, palette } from "../theme";

export interface TabBarTheme {
	tabBarActiveTintColor: string;
	tabBarInactiveTintColor: string;
	tabBarStyle: { backgroundColor: string; borderTopColor: string };
	tabBarLabelStyle: { fontFamily: string; fontWeight: "600"; fontSize: number };
	tabBarBadgeStyle: {
		backgroundColor: string;
		color: string;
		fontFamily: string;
		fontWeight: "600";
	};
}

/** Tab bar screen options for a scheme; navigation takes colours, not classes. */
export function tabBarTheme(scheme: "light" | "dark"): TabBarTheme {
	const colors = palette(scheme);
	return {
		tabBarActiveTintColor: colors.primary,
		tabBarInactiveTintColor: colors.muted,
		tabBarStyle: { backgroundColor: colors.bg, borderTopColor: colors.border },
		tabBarLabelStyle: { fontFamily: fontFamily.text, fontWeight: "600", fontSize: 12 },
		tabBarBadgeStyle: {
			backgroundColor: colors.primary,
			color: colors["on-primary"],
			fontFamily: fontFamily.text,
			fontWeight: "600",
		},
	};
}

/** Profile tab badge for the unread notification count; none at zero. */
export function profileBadge(unread: number): number | string | undefined {
	if (unread <= 0) return undefined;
	return unread > 99 ? "99+" : unread;
}
