import type { ReactElement } from "react";
import { Pressable, Text, View } from "react-native";
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
	const count = t(locale, "account.inbox.unreadCount", { count: unread });
	return (
		<>
			{signedIn ? (
				<>
					<Pressable
						accessibilityRole="button"
						accessibilityLabel={unread > 0 ? `${inbox}, ${count}` : inbox}
						onPress={() => onOpen("/notifications")}
						className="min-h-11 flex-row items-center justify-between gap-3 border-t-2 border-fg pt-4 dark:border-fg-dark"
					>
						<Text className="font-display text-24 font-bold text-fg dark:text-fg-dark">
							{inbox}
						</Text>
						{unread > 0 ? (
							<View
								testID="unread-badge"
								className="min-w-8 items-center rounded-full bg-primary px-2 py-1 dark:bg-primary-dark"
							>
								<Text className="font-text text-14 font-semibold tabular-nums text-on-primary dark:text-on-primary-dark">
									{unread > 99 ? "99+" : unread}
								</Text>
							</View>
						) : null}
					</Pressable>
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
