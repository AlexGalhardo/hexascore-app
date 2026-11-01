import type { ReactElement } from "react";
import { NotificationsScreen } from "../src/notifications/Notifications";
import { notificationsApi } from "../src/notifications/runtime";
import { useLocale } from "../src/providers/root";

export default function NotificationsRoute(): ReactElement {
	return <NotificationsScreen api={notificationsApi} locale={useLocale()} />;
}
