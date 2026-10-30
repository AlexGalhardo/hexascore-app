import { Head, InfiniteScroll, router } from "@inertiajs/react";

import { useLocalized } from "../../lib/useLocalized";
import { useT } from "../../lib/useT";

interface Notification {
	id: string;
	kind: string;
	payload: Record<string, string | number>;
	readAt: string | null;
	createdAt: string;
}

export default function Notifications({
	notifications,
}: {
	notifications: { data: Notification[] };
}) {
	const t = useT();
	const l = useLocalized();
	const rows = notifications.data;

	return (
		<>
			<Head title={t("notifications.title")} />
			<div className="mb-4 flex items-center justify-between gap-3">
				<h1 className="text-2xl font-semibold">{t("notifications.title")}</h1>
				{rows.some((n) => !n.readAt) && (
					<button
						type="button"
						className="font-medium underline"
						onClick={() => router.post(l("/me/notifications/read-all"))}
					>
						{t("notifications.markAllRead")}
					</button>
				)}
			</div>
			{rows.length === 0 ? (
				<p>{t("notifications.empty")}</p>
			) : (
				<InfiniteScroll data="notifications" as="ul" className="divide-y divide-border">
					{rows.map((n) => (
						<li key={n.id} className="flex flex-wrap items-center gap-3 py-3">
							<span
								className={`min-w-0 flex-1 ${n.readAt ? "text-muted" : "font-medium"}`}
							>
								{t(`notifications.kind.${n.kind}`, n.payload)}
								<span className="ml-2 text-sm text-muted">
									{new Date(n.createdAt).toLocaleString()}
								</span>
							</span>
							{!n.readAt && (
								<button
									type="button"
									className="text-sm underline"
									onClick={() => router.post(l(`/me/notifications/${n.id}/read`))}
								>
									{t("notifications.markRead")}
								</button>
							)}
						</li>
					))}
				</InfiniteScroll>
			)}
		</>
	);
}
