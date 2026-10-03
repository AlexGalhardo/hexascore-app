import { isCronAuthorized } from "./cron-auth";
import { env } from "./env";

/** Vercel Cron and the ingestion jobs send `Authorization: Bearer $CRON_SECRET`. */
export function isAuthorizedCron(request: Request): boolean {
	return isCronAuthorized(request.headers.get("authorization"), env.CRON_SECRET);
}
