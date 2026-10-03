import { eq } from "drizzle-orm";
import { z } from "zod";
import type { Database } from "@/lib/db/client";
import { user } from "@/lib/db/schema/auth";
import { lists, reports } from "@/lib/db/schema/social";

import { REPORT_DETAILS_MAX, REPORT_REASONS } from "./constants";

// Users are reported by nickname (the UI never sees user ids), lists by id.
export const reportInputSchema = z.object({
	target: z.discriminatedUnion("type", [
		z.object({ type: z.literal("list"), id: z.uuid() }),
		z.object({ type: z.literal("user"), nickname: z.string().trim().min(1).max(64) }),
	]),
	reason: z.enum(REPORT_REASONS),
	details: z
		.string()
		.trim()
		.max(REPORT_DETAILS_MAX)
		.transform((v) => v || null),
});
export type ReportInput = z.infer<typeof reportInputSchema>;

/** "ok" also covers a duplicate of an open report (deduped by the partial unique index). */
export async function createReport(
	db: Database,
	reporterId: string,
	input: ReportInput,
): Promise<"ok" | "notFound" | "self"> {
	let targetId: string;
	if (input.target.type === "list") {
		const [list] = await db
			.select({ id: lists.id, ownerId: lists.userId })
			.from(lists)
			.where(eq(lists.id, input.target.id))
			.limit(1);
		if (!list) return "notFound";
		if (list.ownerId === reporterId) return "self";
		targetId = list.id;
	} else {
		const [target] = await db
			.select({ id: user.id })
			.from(user)
			.where(eq(user.username, input.target.nickname.toLowerCase()))
			.limit(1);
		if (!target) return "notFound";
		if (target.id === reporterId) return "self";
		targetId = target.id;
	}
	await db
		.insert(reports)
		.values({
			reporterId,
			targetType: input.target.type,
			targetId,
			reason: input.reason,
			details: input.details,
		})
		.onConflictDoNothing();
	return "ok";
}
