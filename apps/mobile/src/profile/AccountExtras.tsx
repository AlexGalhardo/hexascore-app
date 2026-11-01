import type { ReactElement } from "react";
import { type AppLocale, t } from "../i18n";
import { type NotificationsApi, useUnread } from "../notifications/useUnread";
import { Btn } from "../ui/Btn";
import { DeleteAccount, type DeletionWrite } from "./DeleteAccount";
import { ExportData, type ExportWrite } from "./Export";

/** Profile rows below devices: notifications (unread badge), export, deletion and About. */
export function AccountExtras({
	locale,
	signedIn,
	notifications,
	write,
	onScheduled,
	onOpen,
}: {
	locale: AppLocale;
	signedIn: boolean;
	notifications: NotificationsApi;
	write: ExportWrite & DeletionWrite;
	onScheduled: (deletesAt: string) => void;
	onOpen: (path: "/notifications" | "/about") => void;
}): ReactElement {
	const unread = useUnread(notifications, signedIn);
	const inbox = t(locale, "account.inbox.title");
	return (
		<>
			{signedIn ? (
				<>
					<Btn
						variant="secondary"
						label={
							unread > 0
								? `${inbox} · ${t(locale, "account.inbox.unreadCount", { count: unread })}`
								: inbox
						}
						onPress={() => onOpen("/notifications")}
					/>
					<ExportData locale={locale} write={write} />
					<DeleteAccount locale={locale} write={write} onScheduled={onScheduled} />
				</>
			) : null}
			<Btn
				variant="ghost"
				label={t(locale, "account.about.title")}
				onPress={() => onOpen("/about")}
			/>
		</>
	);
}
