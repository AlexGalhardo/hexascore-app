import { appClient } from "../auth/runtime";
import type { NotificationsApi } from "./useUnread";

export const notificationsApi: NotificationsApi = {
	read: appClient.read,
	write: (call) => appClient.write(appClient.newIdempotencyKey(), call),
};
