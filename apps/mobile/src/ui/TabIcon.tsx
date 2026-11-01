import type { ReactElement } from "react";
import { Icon, type IconName } from "./Icon";

/** Tab route -> [unfocused, focused] glyph. Keys follow `app/(tabs)/` file names. */
export const TAB_ICONS = {
	index: ["home-outline", "home"],
	"[type]": ["grid-outline", "grid"],
	search: ["search-outline", "search"],
	lists: ["list-outline", "list"],
	profile: ["person-outline", "person"],
} as const satisfies Record<string, readonly [IconName, IconName]>;

export type TabRoute = keyof typeof TAB_ICONS;

export interface TabIconProps {
	route: TabRoute;
	focused: boolean;
	color: string;
	size: number;
}

/** Use as `tabBarIcon: (props) => <TabIcon route="search" {...props} />`; the tab bar passes the tint. */
export function TabIcon({ route, focused, color, size }: TabIconProps): ReactElement {
	return <Icon name={TAB_ICONS[route][focused ? 1 : 0]} size={size} color={color} />;
}
