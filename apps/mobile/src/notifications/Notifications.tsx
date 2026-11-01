import type { ReactElement } from "react";
import { type AppLocale, t } from "../i18n";
import { NotificationsPage } from "../ui/screens/NotificationsPage";
import { type NotificationsApi, useMarkRead, useNotifications } from "./useUnread";

// Kinds the server sends (SDD 14 §6.5); anything newer falls back to the generic notice.
const KINDS = new Set([
	"suggestion_decided",
	"suggestion_reverted",
	"suggestion_expired",
	"report_resolved",
	"rating_flagged",
	"list_hidden",
]);

export function NotificationsScreen({
	api,
	locale,
}: {
	api: NotificationsApi;
	locale: AppLocale;
}): ReactElement {
	const page = useNotifications(api);
	const markRead = useMarkRead(api);
	const rows = page.data?.data ?? [];

	return (
		<NotificationsPage
			title={t(locale, "account.inbox.title")}
			rows={rows.map((n) => ({
				id: n.id,
				text: t(locale, `account.inbox.kind.${KINDS.has(n.kind) ? n.kind : "system"}`),
				time: n.createdAt ? new Date(n.createdAt).toLocaleDateString(locale) : "",
				unread: n.readAt === null,
			}))}
			labels={{
				unreadCount: (count) => t(locale, "account.inbox.unreadCount", { count }),
				markAll: t(locale, "account.inbox.markAll"),
				unread: t(locale, "account.inbox.unread"),
				empty: t(locale, "account.inbox.empty"),
			}}
			onOpen={(row) => {
				if (row.unread) void markRead([row.id]);
			}}
			onMarkAll={() => void markRead(rows.filter((n) => n.readAt === null).map((n) => n.id))}
		/>
	);
}
