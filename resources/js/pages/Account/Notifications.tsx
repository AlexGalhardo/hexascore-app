import { Head, InfiniteScroll, router } from "@inertiajs/react";

import Btn from "../../components/ui/Btn";
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
								{t(`notifications.kind.${n.kind}`, {
									...n.payload,
									// The server stores the raw target type ("review", "comment"); show it translated.
									...(typeof n.payload.targetType === "string" && {
										targetType: t(
											`notifications.target.${n.payload.targetType}`,
										),
									}),
								})}
								<span className="ml-2 text-sm text-muted">
									{new Date(n.createdAt).toLocaleString()}
								</span>
							</span>
							{n.kind === "follow_request" &&
								typeof n.payload.nickname === "string" && (
									<span className="flex w-full gap-2 sm:w-auto">
										{(["accept", "decline"] as const).map((answer) => (
											<Btn
												key={answer}
												kind={answer === "accept" ? "primary" : "secondary"}
												onClick={() =>
													router.post(
														l(
															`/me/follow-requests/${n.payload.nickname}/${answer}`,
														),
													)
												}
											>
												{t(
													answer === "accept"
														? "notifications.followAccept"
														: "notifications.followDecline",
												)}
											</Btn>
										))}
									</span>
								)}
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
