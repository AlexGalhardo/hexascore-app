import { getFormatter, getTranslations } from "next-intl/server";
import { deletionDate } from "@/lib/account-deletion";
import { getSessionUser } from "@/lib/session";
import { CancelDeletionButton } from "./cancel-deletion-button";

export async function DeletionBanner() {
	const me = await getSessionUser();
	if (!me?.deletionScheduledAt) return null;
	const [t, format] = await Promise.all([getTranslations("account.banner"), getFormatter()]);

	return (
		<div role="status" className="border-b border-score-low bg-surface">
			<div className="mx-auto flex max-w-6xl flex-wrap items-center gap-3 px-4 py-3 text-sm">
				<p className="flex-1">
					{t("text", { date: format.dateTime(deletionDate(me.deletionScheduledAt), { dateStyle: "long" }) })}
				</p>
				<CancelDeletionButton label={t("cancel")} />
			</div>
		</div>
	);
}
