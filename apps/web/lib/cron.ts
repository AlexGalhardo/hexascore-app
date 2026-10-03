import { timingSafeEqual } from "node:crypto";
import { env } from "./env";

/** Vercel Cron and the ingestion jobs send `Authorization: Bearer $CRON_SECRET`. */
export function isAuthorizedCron(request: Request): boolean {
	const secret = env.CRON_SECRET;
	if (!secret) return false;
	const given = Buffer.from(request.headers.get("authorization") ?? "");
	const expected = Buffer.from(`Bearer ${secret}`);
	return given.length === expected.length && timingSafeEqual(given, expected);
}
