import type { ReactElement } from "react";
import { NotificationsScreen } from "../src/notifications/Notifications";
import { notificationsApi } from "../src/notifications/runtime";
import { useLocale } from "../src/providers/root";
import { Screen } from "../src/ui/Screen";

export default function NotificationsRoute(): ReactElement {
	return (
		<Screen>
			<NotificationsScreen api={notificationsApi} locale={useLocale()} />
		</Screen>
	);
}
