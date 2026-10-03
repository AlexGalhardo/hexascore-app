"use server";

import { type ActionResult, fail, ok } from "@/lib/action-result";
import { db } from "@/lib/db/client";
import { asMember } from "@/lib/member";
import { hitRateLimit } from "@/lib/rate-limit";
import { getSessionUser } from "@/lib/session";
import { createReport, reportInputSchema } from "@/lib/social/reports";

export async function report(input: unknown): Promise<ActionResult> {
	const parsed = reportInputSchema.safeParse(input);
	if (!parsed.success) return fail("invalid");
	const member = asMember(await getSessionUser());
	if (typeof member === "string") return fail(member);
	if (!(await hitRateLimit(db, `report:${member.id}`, 10, 3_600_000))) return fail("rateLimited");
	try {
		const result = await createReport(db, member.id, parsed.data);
		if (result === "notFound") return fail("notFound");
		if (result === "self") return fail("forbidden");
	} catch (error) {
		console.error("[reports] create failed:", error);
		return fail("server");
	}
	return ok;
}
