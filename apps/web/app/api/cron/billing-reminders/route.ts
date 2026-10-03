import { createTranslator } from "next-intl";
import { ActionEmail } from "@/emails/action-email";
import { getPathname } from "@/i18n/navigation";
import { claimReminder, dueReminders, expireSubscriptions, releaseReminder } from "@/lib/billing/reminders";
import { APP_NAME, APP_URL } from "@/lib/config";
import { isAuthorizedCron } from "@/lib/cron";
import { db } from "@/lib/db/client";
import { sendEmail } from "@/lib/email";
import messages from "@/messages/pt-BR/billing.json";

// PIX is BRL-only (Brazilian customers) and no locale is stored per user, so reminders go out in pt-BR.
const LOCALE = "pt-BR";

export async function GET(request: Request): Promise<Response> {
	if (!isAuthorizedCron(request)) return new Response("Unauthorized", { status: 401 });
	const now = new Date();
	const t = createTranslator({ locale: LOCALE, messages, namespace: "billing" });
	const url = `${APP_URL}${getPathname({ href: "/profile/api/checkout", locale: LOCALE })}`;
	let sent = 0;
	let failed = 0;
	try {
		for (const r of await dueReminders(db, now)) {
			if (!(await claimReminder(db, r))) continue;
			const plan = t(`plans.${r.plan as "pro" | "business"}`);
			const date = new Intl.DateTimeFormat(LOCALE, { dateStyle: "long", timeZone: "America/Sao_Paulo" }).format(
				r.currentPeriodEnd,
			);
			const subject = t("reminder.subject", { days: r.days, plan });
			try {
				await sendEmail(
					r.email,
					`${APP_NAME}: ${subject}`,
					ActionEmail({
						appName: APP_NAME,
						heading: subject,
						body: t("reminder.body", { plan, date }),
						actionLabel: t("reminder.action"),
						url,
					}),
				);
				sent++;
			} catch (error) {
				failed++;
				console.error(`[cron] billing reminder to ${r.userId} failed`, error);
				await releaseReminder(db, r);
			}
		}
		const expired = await expireSubscriptions(db, now);
		console.info("[cron] billing-reminders", JSON.stringify({ sent, failed, expired }));
		return Response.json({ ok: failed === 0, sent, failed, expired }, { status: failed ? 500 : 200 });
	} catch (error) {
		console.error("[cron] billing-reminders failed", error);
		return Response.json({ ok: false }, { status: 500 });
	}
}
