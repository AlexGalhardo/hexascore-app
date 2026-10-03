import { purgeScheduledAccounts } from "@/lib/account-purge";
import { getStripe } from "@/lib/billing/stripe";
import { isCronAuthorized } from "@/lib/cron-auth";
import { db } from "@/lib/db/client";
import { env } from "@/lib/env";

// Vercel Cron calls this daily (vercel.json) with `Authorization: Bearer $CRON_SECRET`.
export async function GET(request: Request): Promise<Response> {
	if (!isCronAuthorized(request.headers.get("authorization"), env.CRON_SECRET)) {
		return Response.json({ error: "unauthorized" }, { status: 401 });
	}
	try {
		const stripe = getStripe();
		const purged = await purgeScheduledAccounts(
			db,
			new Date(),
			stripe
				? async (id) => {
						await stripe.subscriptions.cancel(id);
					}
				: undefined,
		);
		console.info(`[cron:purge-accounts] anonymized ${purged.length} account(s)`);
		return Response.json({ purged: purged.length });
	} catch (error) {
		console.error("[cron:purge-accounts] failed:", error);
		return Response.json({ error: "purge failed" }, { status: 500 });
	}
}
